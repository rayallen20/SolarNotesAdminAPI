package biz

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"

	"SolarNotesAdminAPI/article_preview"
	"SolarNotesAdminAPI/conf"
	"SolarNotesAdminAPI/model"
	reqArticle "SolarNotesAdminAPI/param/req/v1/article"
	"SolarNotesAdminAPI/sys_error"
)

// Article 保存当前请求已成功读取的文章预览.
type Article struct {
	ID      int64  // ID 是 article 表的主键.
	Title   string // Title 使用文件节点当前显示的名称.
	Content string // Content 保留 Markdown 格式, 供控制器组装响应.

	root       *os.Root // root 限定当前预览读取文件的范围.
	assetBase  *url.URL // assetBase 是从配置解析出的公开资源基址.
	articleDir string   // articleDir 是当前正文所在目录相对于存储根的真实路径.
}

// Preview 从完整请求加载文章, 成功后才更新 ID, Title 和 Content.
func (a *Article) Preview(ctx context.Context, req *reqArticle.PreviewReq) error {
	if a == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Article.Preview: nil receiver")}
	}
	if ctx == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Article.Preview: nil context")}
	}
	if req == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Article.Preview: nil request")}
	}
	if req.Catalogue == nil {
		return &sys_error.InternalError{Cause: errors.New("biz.Article.Preview: nil catalogue reference")}
	}

	cfg := conf.Get()
	if cfg == nil {
		return &sys_error.InternalError{Cause: errors.New("application configuration is not loaded")}
	}
	orm := &model.Article{}
	title, err := orm.LoadForPreview(ctx, req.Catalogue.ID)
	if err != nil {
		switch {
		case errors.Is(err, model.CatalogueNotFoundErr):
			return sys_error.CatalogueNotFoundErr
		case errors.Is(err, model.ArticlePreviewUnavailableErr):
			return sys_error.ArticlePreviewUnavailableErr
		default:
			return &sys_error.DatabaseError{Cause: err}
		}
	}
	content, err := a.readPreviewContent(cfg.Storage.RootDir, cfg.Asset.BaseURL, orm.Path)
	if err != nil {
		if errors.Is(err, sys_error.ArticlePreviewUnavailableErr) {
			return sys_error.ArticlePreviewUnavailableErr
		}
		return &sys_error.InternalError{Cause: err}
	}
	a.ID = orm.ID
	a.Title = title
	a.Content = content
	return nil
}

// readPreviewContent 读取正文, 并把当前文章的图片地址解析方法交给 Markdown 解析器.
func (a *Article) readPreviewContent(rootDir, baseURL, storedPath string) (string, error) {
	err := a.openAssets(rootDir, baseURL)
	if err != nil {
		return "", err
	}
	defer a.closeAssets()

	source, err := a.readMarkdown(rootDir, storedPath)
	if err != nil {
		return "", err
	}
	return article_preview.RewriteImages(source, a.imageURL)
}

// openAssets 在解析文章路径之前打开存储根, 并校验用于拼接的 HTTP(S) 基址.
func (a *Article) openAssets(rootDir, baseURL string) error {
	if !filepath.IsAbs(rootDir) {
		return errors.New("storage.root_dir must be absolute")
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	if (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.User != nil ||
		base.RawQuery != "" || base.ForceQuery || base.Fragment != "" {
		return errors.New("invalid asset.base_url")
	}
	realRoot, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		return err
	}
	root, err := os.OpenRoot(realRoot)
	if err != nil {
		return err
	}
	a.root = root
	a.assetBase = base
	return nil
}

// closeAssets 释放存储根并清理本次预览使用的路径信息.
func (a *Article) closeAssets() {
	_ = a.root.Close()
	a.root = nil
	a.assetBase = nil
	a.articleDir = ""
}

// readMarkdown 读取 UTF-8 正文, 成功时记下文章所在目录相对于存储根的真实路径.
func (a *Article) readMarkdown(rootDir, storedPath string) ([]byte, error) {
	name := storedPath
	if filepath.IsAbs(name) {
		relative, err := filepath.Rel(rootDir, name)
		if err != nil {
			return nil, sys_error.ArticlePreviewUnavailableErr
		}
		name = relative
	}
	if !strings.EqualFold(filepath.Ext(name), ".md") {
		return nil, sys_error.ArticlePreviewUnavailableErr
	}
	actual, err := a.resolveStoragePath(name, false)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(filepath.Ext(actual), ".md") {
		return nil, sys_error.ArticlePreviewUnavailableErr
	}
	// Linux 和 macOS 上用 O_NONBLOCK 打开, 避免在 Stat 识别 FIFO 之前阻塞.
	f, err := a.root.OpenFile(actual, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, sys_error.ArticlePreviewUnavailableErr
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, sys_error.ArticlePreviewUnavailableErr
	}
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(content) {
		return nil, sys_error.ArticlePreviewUnavailableErr
	}
	a.articleDir = filepath.Dir(actual)
	return content, nil
}

// resolveStoragePath 解析符号链接并检查根边界. 普通缺图可以保留尚不存在的尾部路径.
func (a *Article) resolveStoragePath(name string, allowMissing bool) (string, error) {
	if !filepath.IsLocal(name) || strings.ContainsRune(name, 0) {
		return "", sys_error.ArticlePreviewUnavailableErr
	}
	rootDir := a.root.Name()
	candidate := filepath.Join(rootDir, name)
	var missing []string
	for {
		resolved, err := filepath.EvalSymlinks(candidate)
		if err == nil {
			relative, relErr := filepath.Rel(rootDir, resolved)
			if relErr != nil || !filepath.IsLocal(relative) {
				return "", sys_error.ArticlePreviewUnavailableErr
			}
			for i := len(missing) - 1; i >= 0; i-- {
				relative = filepath.Join(relative, missing[i])
			}
			return relative, nil
		}
		if errors.Is(err, syscall.ENOTDIR) || errors.Is(err, syscall.ELOOP) {
			return "", sys_error.ArticlePreviewUnavailableErr
		}
		if !errors.Is(err, fs.ErrNotExist) || candidate == rootDir {
			return "", err
		}
		if !allowMissing {
			return "", sys_error.ArticlePreviewUnavailableErr
		}
		// 失效的符号链接本身仍有目录项, 普通缺失路径则没有.
		_, statErr := os.Lstat(candidate)
		if statErr == nil {
			return "", sys_error.ArticlePreviewUnavailableErr
		}
		if !errors.Is(statErr, fs.ErrNotExist) {
			return "", statErr
		}
		missing = append(missing, filepath.Base(candidate))
		candidate = filepath.Dir(candidate)
	}
}

// imageURL 将文章引用的本地图片映射到该文件的公开 URL.
func (a *Article) imageURL(destination string) (string, error) {
	u, err := url.Parse(destination)
	if err != nil {
		return "", fmt.Errorf("%w: invalid image URI", sys_error.ArticlePreviewUnavailableErr)
	}
	if u.Host != "" && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "") {
		return destination, nil
	}
	if u.Scheme != "" || u.Path == "" || strings.Contains(u.Path, "\\") {
		return "", sys_error.ArticlePreviewUnavailableErr
	}
	name := filepath.FromSlash(u.Path)
	if strings.HasPrefix(u.Path, "/") {
		name = strings.TrimPrefix(name, string(filepath.Separator))
	} else {
		name = filepath.Join(a.articleDir, name)
	}
	if !a.isPublicImagePath(filepath.Clean(name)) {
		return "", sys_error.ArticlePreviewUnavailableErr
	}
	actual, err := a.resolveStoragePath(name, true)
	if err != nil {
		return "", err
	}
	if !a.isPublicImagePath(actual) {
		return "", sys_error.ArticlePreviewUnavailableErr
	}
	path := filepath.ToSlash(actual)
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	result := *a.assetBase
	result.Path = strings.TrimRight(result.Path, "/") + "/" + path
	result.RawPath = strings.TrimRight(a.assetBase.EscapedPath(), "/") + "/" + strings.Join(segments, "/")
	result.RawQuery, result.ForceQuery = u.RawQuery, u.ForceQuery
	result.Fragment, result.RawFragment = u.Fragment, u.RawFragment
	return result.String(), nil
}

// isPublicImagePath 检查 Nginx 对隐藏路径分量和 .md 的限制.
func (a *Article) isPublicImagePath(name string) bool {
	if !filepath.IsLocal(name) || strings.EqualFold(filepath.Ext(name), ".md") {
		return false
	}
	for _, segment := range strings.Split(filepath.ToSlash(name), "/") {
		if strings.HasPrefix(segment, ".") {
			return false
		}
	}
	return true
}
