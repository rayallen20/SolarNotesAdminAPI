package article_preview

import (
	"bytes"
	"errors"
	"html"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// imageEdit 记录一次针对源码字节区间的替换.
type imageEdit struct {
	start, end int    // start 和 end 标记左闭右开的字节范围.
	value      string // value 是新写入的 Markdown 片段.
}

// imageParser 记录 Goldmark 解析图片时消费的源码位置. 每篇文章创建一个实例.
type imageParser struct {
	parser.InlineParser                              // InlineParser 委托给 Goldmark 的标准链接解析器.
	resolve             func(string) (string, error) // resolve 将单个目标转换为公开 URL.
	edits               []imageEdit                  // edits 只收集需要改写的图片.
	err                 error                        // err 保留第一次解析或映射失败的原因.
}

// RewriteImages 重写图片目标, 并按原字节顺序拼回其余正文.
func RewriteImages(source []byte, resolve func(string) (string, error)) (string, error) {
	p := &imageParser{InlineParser: parser.NewLinkParser(), resolve: resolve}
	inlines := parser.DefaultInlineParsers()
	for i, entry := range inlines {
		if entry.Value == p.InlineParser {
			inlines[i].Value = p
		}
	}
	markdown := parser.NewParser(
		parser.WithBlockParsers(parser.DefaultBlockParsers()...),
		parser.WithInlineParsers(inlines...),
		parser.WithParagraphTransformers(parser.DefaultParagraphTransformers()...),
	)
	markdown.Parse(text.NewReader(source))
	if p.err != nil {
		return "", p.err
	}
	sort.Slice(p.edits, func(i, j int) bool { return p.edits[i].start < p.edits[j].start })
	var out strings.Builder
	cursor := 0
	for _, edit := range p.edits {
		if edit.start < cursor || edit.end < edit.start || edit.end > len(source) {
			return "", errors.New("invalid Markdown image range")
		}
		out.Write(source[cursor:edit.start])
		out.WriteString(edit.value)
		cursor = edit.end
	}
	out.Write(source[cursor:])
	return out.String(), nil
}

// CloseBlock 转发标准解析器的收尾操作, 防止链接状态跨段落残留.
func (p *imageParser) CloseBlock(parent ast.Node, r text.Reader, ctx parser.Context) {
	p.InlineParser.(parser.CloseBlocker).CloseBlock(parent, r, ctx)
}

// legacyImage 兼容简单单行图片中带裸空格的文件名. 标准标题和尖括号写法交给 Goldmark.
var legacyImage = regexp.MustCompile(`^!\[[^\[\]\\\x60\r\n]*\]\(([^()<>"'\\\t\r\n ]+(?: +[^()<>"'\\\t\r\n ]+)+)\)`)

// Parse 在解析器处理行内语法时定位图片, 记录真正需要替换的区间.
func (p *imageParser) Parse(parent ast.Node, r text.Reader, ctx parser.Context) ast.Node {
	line, start := r.PeekLine()
	if line[0] == '!' {
		if match := legacyImage.FindSubmatchIndex(line); match != nil {
			p.replace(line[match[2]:match[3]], start.Start+match[2], start.Start+match[3], nil)
			r.Advance(match[1])
			return ast.NewString(line[:match[1]])
		}
	}
	lineNo, saved := r.Position()
	node := p.InlineParser.Parse(parent, r, ctx)
	img, ok := node.(*ast.Image)
	if !ok || len(img.Destination) == 0 {
		return node
	}
	endLine, end := r.Position()
	afterLabel := start.Start + 1
	if len(line) > 1 && line[1] == '(' && end.Start > afterLabel {
		// 只定位目标, 保留括号, 标题及其间的排版.
		r.SetPosition(lineNo, saved)
		r.Advance(2)
		r.SkipSpaces()
		_, dest := r.Position()
		from := dest.Start
		if r.Peek() == '<' {
			from++
		}
		r.SetPosition(endLine, end)
		to := from + len(img.Destination)
		if from < 0 || to > len(r.Source()) || !bytes.Equal(r.Source()[from:to], img.Destination) {
			p.err = errors.New("cannot locate Markdown image destination")
			return node
		}
		p.replace(img.Destination, from, to, nil)
	} else {
		// 把这次图片引用展开为行内图片, 普通链接可以继续使用同一份引用定义.
		p.replace(img.Destination, afterLabel, end.Start, img)
	}
	return node
}

// markdownEscape 匹配单个反斜杠标点转义或带分号的字符实体.
var markdownEscape = regexp.MustCompile(`\\[[:punct:]]|&#(?:[0-9]{1,7}|[xX][0-9a-fA-F]{1,6});|&[A-Za-z][A-Za-z0-9]*;`)

// markdownValue 一次解开原文中的转义和实体, 保留 URL 百分号编码及转义后产生的字面文本.
func markdownValue(raw []byte) string {
	return markdownEscape.ReplaceAllStringFunc(string(raw), decodeMarkdownEscape)
}

// decodeMarkdownEscape 保留未知的实体名称, 数字实体按十进制或显式十六进制解码.
func decodeMarkdownEscape(token string) string {
	if token[0] == '\\' {
		return token[1:]
	}
	if token[1] == '#' {
		return html.UnescapeString(token)
	}
	if entity, ok := util.LookUpHTML5EntityByName(token[1 : len(token)-1]); ok {
		return string(entity.Characters)
	}
	return token
}

// urlSyntax 保护目标中的 Markdown 定界符和字符实体.
var urlSyntax = strings.NewReplacer("(", "%28", ")", "%29", "<", "%3C", ">", "%3E",
	" ", "%20", "\\", "%5C", "\r", "%0D", "\n", "%0A", "&", "&amp;")

// replace 仅在目标变化时记录替换, 外部 HTTP(S) 图片可完整保留原文.
func (p *imageParser) replace(raw []byte, start, end int, ref *ast.Image) {
	if p.err != nil {
		return
	}
	value := markdownValue(raw)
	target, err := p.resolve(value)
	if err != nil {
		p.err = err
		return
	}
	if target == value {
		return
	}
	target = urlSyntax.Replace(target)
	if ref != nil {
		if len(ref.Title) > 0 {
			title := html.EscapeString(markdownValue(ref.Title))
			title = strings.NewReplacer("\\", "\\\\", "\r", "&#13;", "\n", "&#10;").Replace(title)
			target += ` "` + title + `"`
		}
		target = "(" + target + ")"
	}
	p.edits = append(p.edits, imageEdit{start: start, end: end, value: target})
}
