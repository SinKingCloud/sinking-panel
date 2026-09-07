package server

import (
	"strconv"
	"time"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver"
)

const defaultPageStyle = `<style>
        :root {
            color-scheme: light dark;
            --page: #e9e9e9;
            --text: #171717;
            --text-secondary: rgba(0, 0, 0, .58);
            --text-tertiary: rgba(0, 0, 0, .4);
        }
        * {
            box-sizing: border-box;
            letter-spacing: 0;
        }
        html,
        body {
            min-height: 100%;
        }
        html {
            background: var(--page);
        }
        body {
            min-width: 280px;
            min-height: 100vh;
            min-height: 100dvh;
            margin: 0;
            display: grid;
            grid-template-rows: minmax(0, 1fr) auto;
            background: var(--page);
            color: var(--text);
            font: 14px/1.5715 Arial, "Microsoft YaHei", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
        }
        .stage {
            min-width: 0;
            display: grid;
            place-items: center;
            padding: 48px 24px;
        }
        .content {
            width: 100%;
            max-width: 900px;
            text-align: center;
        }
        .title {
            margin: 0;
            color: var(--text);
            font-size: 64px;
            font-weight: 300;
            line-height: 1.25;
            overflow-wrap: break-word;
        }
        .description {
            margin: 18px 0 0;
            display: flex;
            align-items: center;
            justify-content: center;
            flex-wrap: wrap;
            gap: 8px;
            color: var(--text-secondary);
            font-size: 14px;
            line-height: 24px;
            overflow-wrap: anywhere;
        }
        .status {
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: 12px;
        }
        .divider {
            color: var(--text-tertiary);
        }
        .footer {
            padding: 20px 20px calc(20px + env(safe-area-inset-bottom));
            color: var(--text-tertiary);
            font-size: 12px;
            line-height: 20px;
            text-align: center;
        }
        @media (prefers-color-scheme: dark) {
            :root {
                --page: #171717;
                --text: #f5f5f5;
                --text-secondary: rgba(255, 255, 255, .66);
                --text-tertiary: rgba(255, 255, 255, .42);
            }
        }
        @media (max-width: 640px) {
            .stage {
                padding: 36px 20px;
            }
            .title {
                font-size: 38px;
                line-height: 1.3;
            }
            .description {
                margin-top: 12px;
                gap: 6px;
                font-size: 13px;
                line-height: 22px;
            }
            .footer {
                padding-top: 16px;
            }
        }
        @media (max-height: 420px) {
            .stage {
                place-items: start center;
                padding-top: 32px;
                padding-bottom: 24px;
            }
        }
    </style>` // 默认页面样式

var (
	pageYear            = strconv.Itoa(time.Now().Year())
	defaultNotFoundPage = `<!doctype html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex, nofollow">
    <title>页面不存在</title>` + defaultPageStyle + `
</head>
<body>
    <main class="stage">
        <section class="content" aria-labelledby="page-title">
            <h1 class="title" id="page-title">页面不存在</h1>
            <p class="description"><span class="status">HTTP 404</span><span class="divider" aria-hidden="true">·</span><span>您访问的页面不存在，可能已被移动或删除。</span></p>
        </section>
    </main>
    <footer class="footer">© ` + pageYear + ` All Rights Reserved</footer>
</body>
</html>` // 默认 404 页面
	defaultSiteNotFoundPage = `<!doctype html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex, nofollow">
    <title>网站不存在</title>` + defaultPageStyle + `
</head>
<body>
    <main class="stage">
        <section class="content" aria-labelledby="page-title">
            <h1 class="title" id="page-title">网站不存在</h1>
            <p class="description"><span class="status">HTTP 404</span><span class="divider" aria-hidden="true">·</span><span>当前域名尚未绑定可用的网站。</span></p>
        </section>
    </main>
    <footer class="footer">© ` + pageYear + ` All Rights Reserved</footer>
</body>
</html>` // 默认网站不存在页面
	defaultSiteDisabledPage = `<!doctype html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex, nofollow">
    <title>网站已停用</title>` + defaultPageStyle + `
</head>
<body>
    <main class="stage">
        <section class="content" aria-labelledby="page-title">
            <h1 class="title" id="page-title">网站已停用</h1>
            <p class="description"><span class="status">HTTP 503</span><span class="divider" aria-hidden="true">·</span><span>当前网站已暂停访问，请稍后再试。</span></p>
        </section>
    </main>
    <footer class="footer">© ` + pageYear + ` All Rights Reserved</footer>
</body>
</html>` // 默认网站停用页面
	defaultWAFBlockPage = `<!doctype html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex, nofollow">
    <title>访问被拦截</title>` + defaultPageStyle + `
</head>
<body>
    <main class="stage">
        <section class="content" aria-labelledby="page-title">
            <h1 class="title" id="page-title">访问被拦截</h1>
            <p class="description"><span class="status">WAF</span><span class="divider" aria-hidden="true">·</span><span>当前请求未通过网站安全检查，请联系网站管理员。</span></p>
        </section>
    </main>
    <footer class="footer">© ` + pageYear + ` All Rights Reserved</footer>
</body>
</html>` // 默认 WAF 拦截页面
)

var browseTemplate = `{{- $nonce := uuidv4 -}}
{{- $csp := printf "default-src 'none'; style-src 'nonce-%s'; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'" $nonce -}}
{{- .RespHeader.Set "Content-Security-Policy" $csp -}}
<!doctype html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex, nofollow">
    <title>目录浏览 - {{html .Req.URL.Path}}</title>
    <style nonce="{{$nonce}}">
        :root {
            color-scheme: light dark;
            --page: #e9e9e9;
            --text: #171717;
            --text-secondary: rgba(0, 0, 0, .62);
            --text-tertiary: rgba(0, 0, 0, .42);
            --line: rgba(0, 0, 0, .12);
            --hover: rgba(255, 255, 255, .42);
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --page: #171717;
                --text: #f5f5f5;
                --text-secondary: rgba(255, 255, 255, .66);
                --text-tertiary: rgba(255, 255, 255, .42);
                --line: rgba(255, 255, 255, .14);
                --hover: rgba(255, 255, 255, .06);
            }
        }

        * {
            box-sizing: border-box;
            letter-spacing: 0;
        }

        html,
        body {
            min-height: 100%;
        }

        html {
            background: var(--page);
        }

        body {
            min-width: 280px;
            min-height: 100vh;
            min-height: 100dvh;
            margin: 0;
            display: flex;
            flex-direction: column;
            background: var(--page);
            color: var(--text);
            font: 14px/1.5715 Arial, "Microsoft YaHei", -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
        }

        a {
            color: inherit;
            text-decoration: none;
        }

        svg {
            display: block;
        }

        .page {
            width: 100%;
            max-width: 1120px;
            margin: 0 auto;
            padding: 52px 24px 32px;
            flex: 1;
        }

        .navigation {
            min-height: 34px;
            display: flex;
            align-items: center;
        }

        .back {
            display: inline-flex;
            align-items: center;
            gap: 7px;
            padding: 6px 0;
            color: var(--text-secondary);
            font-size: 13px;
            line-height: 22px;
            transition: color .16s ease;
        }

        .back:hover {
            color: var(--text);
        }

        .back svg {
            width: 15px;
            height: 15px;
            fill: none;
            stroke: currentColor;
            stroke-width: 1.8;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .hero {
            padding: 30px 0 0;
            text-align: center;
        }

        .title {
            margin: 0;
            color: var(--text);
            font-size: 56px;
            font-weight: 300;
            line-height: 1.2;
            overflow-wrap: anywhere;
        }

        .path-scroll {
            max-width: 100%;
            margin: 16px auto 0;
            overflow-x: auto;
            overflow-y: hidden;
            scrollbar-width: none;
            -webkit-overflow-scrolling: touch;
        }

        .path-scroll::-webkit-scrollbar {
            display: none;
        }

        .breadcrumb {
            width: max-content;
            min-width: 100%;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 7px;
            color: var(--text-secondary);
            font-size: 14px;
            font-weight: 400;
            line-height: 24px;
            white-space: nowrap;
        }

        .breadcrumb a,
        .breadcrumb span {
            flex: none;
        }

        .breadcrumb a {
            transition: color .16s ease;
        }

        .breadcrumb a:hover {
            color: var(--text);
        }

        .separator {
            color: var(--text-tertiary);
        }

        .summary {
            margin: 8px 0 0;
            color: var(--text-tertiary);
            font-size: 12px;
            line-height: 20px;
        }

        .summary strong {
            color: var(--text-secondary);
            font-weight: 600;
        }

        .listing {
            margin-top: 44px;
            border-top: 1px solid var(--line);
            border-bottom: 1px solid var(--line);
        }

        table {
            width: 100%;
            border-collapse: collapse;
            table-layout: fixed;
        }

        th,
        td {
            border-bottom: 1px solid var(--line);
            text-align: left;
            vertical-align: middle;
        }

        th {
            height: 44px;
            padding: 8px 12px;
            color: var(--text-tertiary);
            font-size: 12px;
            font-weight: 400;
        }

        td {
            height: 52px;
            padding: 8px 12px;
            color: var(--text-secondary);
            font-size: 13px;
        }

        tbody tr:last-child td {
            border-bottom: 0;
        }

        tbody tr {
            transition: background-color .16s ease;
        }

        tbody tr:hover {
            background: var(--hover);
        }

        .name-column {
            width: auto;
        }

        .modified-column {
            width: 190px;
        }

        .size-column {
            width: 110px;
            text-align: right;
        }

        .sort-link {
            display: inline-flex;
            align-items: center;
            gap: 5px;
            transition: color .16s ease;
        }

        .sort-link:hover,
        .sort-link.active {
            color: var(--text);
        }

        .sort-direction {
            font-size: 10px;
            line-height: 1;
        }

        .name-link {
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 9px;
            color: var(--text);
        }

        .name-link:hover .file-name {
            text-decoration: underline;
            text-underline-offset: 3px;
        }

        .file-icon {
            width: 18px;
            height: 18px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            flex: none;
            color: var(--text-tertiary);
        }

        .file-icon.directory {
            color: var(--text-secondary);
        }

        .file-icon svg {
            width: 16px;
            height: 16px;
            fill: none;
            stroke: currentColor;
            stroke-width: 1.65;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .file-name {
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        .symlink {
            margin-left: 3px;
            color: var(--text-tertiary);
            font-size: 11px;
        }

        .modified,
        .size {
            color: var(--text-tertiary);
            font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
            font-size: 12px;
            white-space: nowrap;
        }

        .size {
            text-align: right;
        }

        .empty-state {
            min-height: 190px;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: 10px;
            color: var(--text-tertiary);
        }

        .empty-state svg {
            width: 26px;
            height: 26px;
            fill: none;
            stroke: currentColor;
            stroke-width: 1.4;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .back:focus-visible,
        .breadcrumb a:focus-visible,
        .name-link:focus-visible,
        .sort-link:focus-visible {
            outline: 1px solid currentColor;
            outline-offset: 4px;
        }

        .footer {
            padding: 18px 20px calc(18px + env(safe-area-inset-bottom));
            color: var(--text-tertiary);
            font-size: 12px;
            line-height: 20px;
            text-align: center;
        }

        @media (max-width: 640px) {
            .page {
                padding: 28px 16px 24px;
            }

            .navigation {
                min-height: 32px;
            }

            .hero {
                padding-top: 22px;
            }

            .title {
                font-size: 38px;
            }

            .path-scroll {
                margin-top: 12px;
            }

            .breadcrumb {
                justify-content: flex-start;
                font-size: 13px;
            }

            .listing {
                margin-top: 30px;
            }

            th,
            td {
                padding-right: 8px;
                padding-left: 8px;
            }

            .modified-column,
            .modified {
                display: none;
            }

            .size-column {
                width: 78px;
            }

            .file-name {
                white-space: nowrap;
            }

            .footer {
                padding-top: 14px;
            }
        }
    </style>
</head>
<body>
    <main class="page" aria-labelledby="directory-title">
        <div class="navigation">
            {{- if .CanGoUp}}
            <a class="back" href=".." aria-label="返回上级目录">
                <svg viewBox="0 0 24 24" aria-hidden="true"><path d="m14 18-6-6 6-6"></path></svg>
                返回上级
            </a>
            {{- end}}
        </div>

        <header class="hero">
            <h1 class="title" id="directory-title">目录浏览</h1>
            <div class="path-scroll">
                <nav class="breadcrumb" aria-label="当前路径">
                    {{- range $index, $crumb := .Breadcrumbs}}
                    {{- if ne $index 0}}<span class="separator">›</span>{{end}}
                    <a href="{{html $crumb.Link}}">{{html $crumb.Text}}</a>
                    {{- end}}
                </nav>
            </div>
            <p class="summary">
                <strong>{{.NumDirs}}</strong> 个目录&nbsp;&nbsp;<strong>{{.NumFiles}}</strong> 个文件
                {{- if gt .NumFiles 0}}&nbsp;&nbsp;{{.HumanTotalFileSize}}{{end}}
            </p>
        </header>

        <section class="listing" aria-label="目录内容">
            {{- if .Items}}
            <table aria-label="目录文件列表">
                <colgroup>
                    <col class="name-column">
                    <col class="modified-column">
                    <col class="size-column">
                </colgroup>
                <thead>
                    <tr>
                        <th>
                            <a class="sort-link{{if eq .Sort "namedirfirst"}} active{{end}}" href="?sort=namedirfirst&amp;order={{if and (eq .Sort "namedirfirst") (eq .Order "asc")}}desc{{else}}asc{{end}}">
                                名称
                                {{- if eq .Sort "namedirfirst"}}<span class="sort-direction">{{if eq .Order "asc"}}↑{{else}}↓{{end}}</span>{{end}}
                            </a>
                        </th>
                        <th class="modified-column">
                            <a class="sort-link{{if eq .Sort "time"}} active{{end}}" href="?sort=time&amp;order={{if and (eq .Sort "time") (eq .Order "asc")}}desc{{else}}asc{{end}}">
                                修改时间
                                {{- if eq .Sort "time"}}<span class="sort-direction">{{if eq .Order "asc"}}↑{{else}}↓{{end}}</span>{{end}}
                            </a>
                        </th>
                        <th class="size-column">
                            <a class="sort-link{{if eq .Sort "size"}} active{{end}}" href="?sort=size&amp;order={{if and (eq .Sort "size") (eq .Order "asc")}}desc{{else}}asc{{end}}">
                                大小
                                {{- if eq .Sort "size"}}<span class="sort-direction">{{if eq .Order "asc"}}↑{{else}}↓{{end}}</span>{{end}}
                            </a>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {{- range .Items}}
                    <tr>
                        <td>
                            <a class="name-link" href="{{html .URL}}" title="{{html .Name}}">
                                {{- if .IsDir}}
                                <span class="file-icon directory" aria-hidden="true">
                                    <svg viewBox="0 0 24 24"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.7-.9l-.8-1.2A2 2 0 0 0 7.9 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"></path></svg>
                                </span>
                                {{- else}}
                                <span class="file-icon" aria-hidden="true">
                                    <svg viewBox="0 0 24 24"><path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"></path><path d="M14 2v6h6"></path></svg>
                                </span>
                                {{- end}}
                                <span class="file-name">{{html .Name}}{{if .IsSymlink}}<span class="symlink">↗</span>{{end}}</span>
                            </a>
                        </td>
                        <td class="modified">{{.HumanModTime "2006-01-02 15:04"}}</td>
                        <td class="size">{{if .IsDir}}-{{else}}{{.HumanSize}}{{end}}</td>
                    </tr>
                    {{- end}}
                </tbody>
            </table>
            {{- else}}
            <div class="empty-state">
                <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.7-.9l-.8-1.2A2 2 0 0 0 7.9 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"></path></svg>
                暂无文件
            </div>
            {{- end}}
        </section>
    </main>
    <footer class="footer">© ` + pageYear + ` All Rights Reserved</footer>
</body>
</html>`

func init() {
	fileserver.BrowseTemplate = browseTemplate
}
