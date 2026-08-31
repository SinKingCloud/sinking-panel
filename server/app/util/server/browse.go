package server

import "github.com/caddyserver/caddy/v2/modules/caddyhttp/fileserver"

const browseTemplate = `{{- $nonce := uuidv4 -}}
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
            --page: #f5f5f5;
            --surface: #ffffff;
            --surface-hover: rgba(0, 0, 0, .04);
            --surface-active: #e6f2ff;
            --text: rgba(0, 0, 0, .88);
            --text-secondary: rgba(0, 0, 0, .65);
            --text-tertiary: rgba(0, 0, 0, .45);
            --border: #f0f0f0;
            --split: rgba(5, 5, 5, .06);
            --primary: #0053fb;
            --primary-hover: #2977ff;
            --radius: 8px;
            --radius-small: 5px;
            --shadow: 0 1px 2px rgba(0, 0, 0, .05), 0 1px 6px -1px rgba(0, 0, 0, .03), 0 2px 4px rgba(0, 0, 0, .03);
        }

        @media (prefers-color-scheme: dark) {
            :root {
                --page: #000000;
                --surface: #141414;
                --surface-hover: rgba(255, 255, 255, .08);
                --surface-active: #0e2759;
                --text: rgba(255, 255, 255, .85);
                --text-secondary: rgba(255, 255, 255, .65);
                --text-tertiary: rgba(255, 255, 255, .45);
                --border: #303030;
                --split: rgba(253, 253, 253, .12);
                --primary: #276de8;
                --primary-hover: #4b86ed;
                --shadow: none;
            }
        }

        * {
            box-sizing: border-box;
        }

        html,
        body {
            min-height: 100%;
        }

        html {
            background: var(--page);
        }

        body {
            margin: 0;
            background: var(--page);
            color: var(--text);
            font: 14px/1.5715 -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
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
            max-width: 1450px;
            margin: 0 auto;
            padding: 10px;
        }

        .workspace {
            min-width: 0;
            overflow: hidden;
            border-radius: var(--radius);
            background: var(--surface);
            box-shadow: var(--shadow);
        }

        .toolbar {
            min-height: 66px;
            padding: 12px 14px;
            display: flex;
            align-items: center;
            gap: 10px;
        }

        .back {
            width: 34px;
            height: 34px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            flex: none;
            border-radius: var(--radius-small);
            background: var(--surface-hover);
            color: var(--text-secondary);
            transition: color .16s ease, background-color .16s ease;
        }

        .back:hover {
            background: var(--surface-active);
            color: var(--primary);
        }

        .back svg {
            fill: none;
            stroke: currentColor;
            stroke-width: 1.8;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        .back:focus-visible,
        .breadcrumb a:focus-visible,
        .name-link:focus-visible,
        .sort-link:focus-visible {
            outline: 2px solid color-mix(in srgb, var(--primary), transparent 58%);
            outline-offset: 2px;
        }

        .heading {
            min-width: 0;
            flex: 1;
        }

        .label {
            margin: 0 0 2px;
            color: var(--text-tertiary);
            font-size: 11px;
            font-weight: 400;
            line-height: 18px;
        }

        .breadcrumb {
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 5px;
            overflow-x: auto;
            overflow-y: hidden;
            color: var(--text-secondary);
            font-size: 14px;
            font-weight: 500;
            line-height: 22px;
            scrollbar-width: none;
            -webkit-overflow-scrolling: touch;
        }

        .breadcrumb::-webkit-scrollbar {
            display: none;
        }

        .breadcrumb a,
        .breadcrumb span {
            flex: none;
        }

        .breadcrumb a:hover {
            color: var(--primary);
        }

        .breadcrumb .separator {
            color: var(--text-tertiary);
            font-weight: 400;
        }

        .summary {
            flex: none;
            color: var(--text-tertiary);
            font-size: 12px;
            white-space: nowrap;
        }

        .summary strong {
            color: var(--text-secondary);
            font-weight: 500;
        }

        .listing {
            margin: 0 10px 10px;
            overflow: hidden;
            border: 1px solid var(--border);
            border-radius: var(--radius);
        }

        table {
            width: 100%;
            border-collapse: collapse;
            table-layout: fixed;
        }

        th,
        td {
            border-bottom: 1px solid var(--split);
            text-align: left;
            vertical-align: middle;
        }

        th {
            height: 42px;
            padding: 8px 12px;
            background: var(--surface-hover);
            color: var(--text-secondary);
            font-size: 12px;
            font-weight: 500;
        }

        td {
            height: 48px;
            padding: 8px 12px;
            color: var(--text-secondary);
            font-size: 13px;
        }

        tbody tr:last-child td {
            border-bottom: 0;
        }

        tbody tr:hover td {
            background: var(--surface-hover);
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
            color: var(--primary);
        }

        .sort-direction {
            font-size: 10px;
            line-height: 1;
        }

        .name-link {
            min-width: 0;
            display: flex;
            align-items: center;
            gap: 8px;
            color: var(--text);
        }

        .name-link:hover {
            color: var(--primary);
        }

        .file-icon {
            width: 20px;
            height: 20px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            flex: none;
            color: var(--text-tertiary);
        }

        .file-icon.directory {
            color: var(--primary);
        }

        .file-icon svg {
            width: 15px;
            height: 15px;
            fill: none;
            stroke: currentColor;
            stroke-width: 1.8;
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
            margin-left: 2px;
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
            gap: 8px;
            color: var(--text-tertiary);
        }

        .empty-state svg {
            width: 28px;
            height: 28px;
            fill: none;
            stroke: currentColor;
            stroke-width: 1.5;
            stroke-linecap: round;
            stroke-linejoin: round;
        }

        @media (max-width: 640px) {
            .page {
                padding: 0;
            }

            .workspace {
                min-height: 100vh;
                border-radius: 0;
                box-shadow: none;
            }

            .toolbar {
                min-height: 60px;
                padding: 10px 12px;
            }

            .back {
                width: 36px;
                height: 36px;
            }

            .label {
                margin-bottom: 0;
                font-size: 10px;
            }

            .breadcrumb {
                font-size: 13px;
            }

            .summary {
                display: none;
            }

            .listing {
                margin: 0 10px 10px;
                border-radius: var(--radius-small);
            }

            th,
            td {
                padding-right: 10px;
                padding-left: 10px;
            }

            .modified-column,
            .modified {
                display: none;
            }

            .size-column {
                width: 88px;
            }

            .file-name {
                white-space: normal;
                display: -webkit-box;
                overflow: hidden;
                -webkit-box-orient: vertical;
                -webkit-line-clamp: 2;
                overflow-wrap: anywhere;
            }
        }
    </style>
</head>
<body>
    <main class="page">
        <section class="workspace" aria-labelledby="directory-title">
            <header class="toolbar">
                {{- if .CanGoUp}}
                <a class="back" href=".." aria-label="返回上级目录" title="返回上级目录">
                    <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true">
                        <path d="m12 19-7-7 7-7"></path>
                        <path d="M19 12H5"></path>
                    </svg>
                </a>
                {{- end}}
                <div class="heading">
                    <h1 class="label" id="directory-title">目录浏览</h1>
                    <nav class="breadcrumb" aria-label="当前路径">
                        {{- range $index, $crumb := .Breadcrumbs}}
                        <a href="{{html $crumb.Link}}">{{html $crumb.Text}}</a>
                        {{- if ne $index 0}}<span class="separator">/</span>{{end}}
                        {{- end}}
                    </nav>
                </div>
                <div class="summary">
                    <strong>{{.NumDirs}}</strong> 个目录&nbsp;&nbsp;<strong>{{.NumFiles}}</strong> 个文件
                    {{- if gt .NumFiles 0}}&nbsp;&nbsp;{{.HumanTotalFileSize}}{{end}}
                </div>
            </header>

            <div class="listing">
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
                                        <svg viewBox="0 0 24 24">
                                            <path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.7-.9l-.8-1.2A2 2 0 0 0 7.9 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"></path>
                                        </svg>
                                    </span>
                                    {{- else}}
                                    <span class="file-icon" aria-hidden="true">
                                        <svg viewBox="0 0 24 24">
                                            <path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"></path>
                                            <path d="M14 2v6h6"></path>
                                        </svg>
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
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                        <path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.7-.9l-.8-1.2A2 2 0 0 0 7.9 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"></path>
                    </svg>
                    暂无文件
                </div>
                {{- end}}
            </div>
        </section>
    </main>
</body>
</html>`

func init() {
	fileserver.BrowseTemplate = browseTemplate
}
