package sys_error

import "errors"

// ArticlePreviewUnavailableErr 表示所选节点当前无法提供 Markdown 预览.
var ArticlePreviewUnavailableErr = errors.New("文章暂时无法预览")
