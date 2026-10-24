from pathlib import Path
from html import escape
import json

from bs4 import BeautifulSoup
from markdown_it import MarkdownIt

root = Path(__file__).resolve().parent.parent
source = root / 'research' / 'claude-ban-fingerprint-research.md'
output = root / 'docs' / 'claude-ban-fingerprint-research.html'
content = MarkdownIt('commonmark', {'html': False}).enable('table').render(source.read_text(encoding='utf-8'))
soup = BeautifulSoup(content, 'html.parser')
navigation = []
for index, heading in enumerate(soup.select('h2'), 1):
    heading['id'] = f'section-{index}'
    navigation.append(f'<a href="#section-{index}">{escape(heading.get_text())}</a>')
for table in soup.select('table'):
    wrapper = soup.new_tag('div', attrs={'class': 'table-scroll', 'tabindex': '0', 'role': 'region', 'aria-label': '资料表格，可横向滚动'})
    table.wrap(wrapper)
    for th in table.select('thead th'):
        th['scope'] = 'col'
for link in soup.select('a[href]'):
    if link['href'].startswith(('https://', 'http://')):
        link['target'] = '_blank'
        link['rel'] = 'noopener noreferrer'
local_files = {
    'claude-ban-fingerprint-sources.json': '../research/claude-ban-fingerprint-sources.json',
}
for code in soup.select('code'):
    if code.get_text() in local_files:
        link = soup.new_tag('a', href=local_files[code.get_text()])
        code.wrap(link)

css = '''
:root{color-scheme:light;--ink:#202e3b;--muted:#637383;--accent:#1a617b;--line:#dce4e7;--paper:#fff;--bg:#f3f5f4}
*{box-sizing:border-box}html{scroll-behavior:smooth;scroll-padding-top:28px}
body{margin:0;background:var(--bg);color:var(--ink);font:16px/1.85 system-ui,-apple-system,"Segoe UI","Microsoft YaHei",sans-serif}
a{color:var(--accent);text-decoration:none;text-underline-offset:3px}a:hover{text-decoration:underline}
a:focus-visible,[tabindex]:focus-visible{outline:3px solid #d38a32;outline-offset:4px}
.skip{position:fixed;left:16px;top:-90px;background:#fff;padding:8px 16px;z-index:5}.skip:focus{top:12px}
.layout{max-width:1510px;margin:auto;padding:40px 32px 72px;display:grid;grid-template-columns:245px minmax(0,1fr);gap:38px}
aside{position:sticky;top:32px;align-self:start;max-height:calc(100vh - 64px);overflow:auto}
.brand{font-size:13px;font-weight:750;letter-spacing:.12em;color:var(--accent);margin-bottom:8px}
.side-title{font-size:20px;font-weight:700;line-height:1.5;margin-bottom:24px}
.nav-title{font-size:12px;color:var(--muted);margin-bottom:12px}
nav{display:grid;gap:5px}nav a{font-size:13px;line-height:1.6;padding:8px 12px;border-left:2px solid var(--line);color:#4d6170}
nav a:hover{background:#e7eeee;border-color:var(--accent);text-decoration:none}
.side-note{font-size:12px;color:var(--muted);margin-top:26px;padding:15px 12px;border-top:1px solid var(--line)}
main{min-width:0;background:var(--paper);border:1px solid var(--line);border-radius:16px;padding:42px 44px;box-shadow:0 10px 45px #243b4b06}
.eyebrow{color:var(--accent);font-size:12px;font-weight:750;letter-spacing:.14em;margin-bottom:12px}
h1{font-size:clamp(26px,3vw,38px);line-height:1.35;letter-spacing:-.035em;margin:0 0 22px}
h2{font-size:23px;line-height:1.5;margin:44px 0 18px;padding-top:25px;border-top:1px solid var(--line)}
p{margin:14px 0}h1+p{font-size:14px;color:var(--muted)}
ul,ol{padding-left:1.5em}li{padding-left:3px;margin:9px 0}strong{font-weight:700}
code{font-family:ui-monospace,Consolas,monospace;font-size:.88em;background:#f0f4f5;border-radius:4px;padding:2px 5px;overflow-wrap:anywhere}
.table-scroll{overflow-x:auto;margin:20px 0 28px;border:1px solid var(--line);border-radius:9px}
table{width:100%;border-collapse:collapse;font-size:14px;line-height:1.8;min-width:620px}
th,td{text-align:left;vertical-align:top;padding:14px 17px;border-bottom:1px solid var(--line);overflow-wrap:anywhere}
th{background:#eaf1f2;color:#35525e;font-size:12px;letter-spacing:.02em}
tr:last-child td{border-bottom:0}tbody tr:nth-child(even){background:#fafcfc}
td:first-child{min-width:52px}td:nth-child(2){min-width:210px}td:last-child{min-width:300px}
td a{font-weight:600}.footer{font-size:12px;color:var(--muted);margin-top:36px;padding-top:20px;border-top:1px solid var(--line)}
@media(max-width:1000px){.layout{grid-template-columns:190px minmax(0,1fr);gap:22px;padding:24px 20px}main{padding:30px 26px}}
@media(max-width:760px){.layout{display:block;padding:16px 12px 36px}aside{position:static;max-height:none;margin:0 4px 20px}.side-title{margin-bottom:14px}.nav-title,.side-note{display:none}nav{display:flex;overflow:auto;gap:8px;padding-bottom:8px}nav a{flex:0 0 165px;border:1px solid var(--line);border-radius:6px;background:white;font-size:12px}main{padding:26px 20px;border-radius:10px}h2{font-size:20px}body{font-size:15px}}
@media(prefers-reduced-motion:reduce){html{scroll-behavior:auto}}
@media print{@page{size:A4;margin:15mm}body{background:white;font-size:10pt}.layout{display:block;max-width:none;padding:0}aside,.skip,.eyebrow{display:none}main{border:0;border-radius:0;box-shadow:none;padding:0}h1{font-size:23pt}h2{font-size:15pt;break-after:avoid}table{min-width:0;font-size:9pt}.table-scroll{overflow:visible}th,td{min-width:0!important;padding:7px 9px}tr{break-inside:avoid}thead{display:table-header-group}a{color:#164f66}}
'''
page = f'''<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="description" content="Claude 封号、设备指纹与遥测研究目录：55 项官方文档、逆向研究、项目、文章与讨论，附证据边界。">
<title>Claude 封号、指纹与遥测 · 资料索引</title>
<style>{css}</style>
</head>
<body>
<a class="skip" href="#content">跳转到正文</a>
<div class="layout">
<aside aria-label="文档导航">
<div class="brand">CLAUDE / RESEARCH</div>
<div class="side-title">封号、指纹与遥测<br>资料索引</div>
<div class="nav-title">目录 · 55 项资料 / 7 类来源</div>
<nav aria-label="章节目录">{''.join(navigation)}</nav>
<div class="side-note">检索于 2026 年 10 月 18 日<br>Asia/Shanghai<br><br>来源、事实与推测分开阅读。每项资料均保留核验说明。</div>
</aside>
<main id="content">
<div class="eyebrow">PUBLIC SOURCES · ANNOTATED RESEARCH</div>
{soup}
<div class="footer">55 项资料 · 7 类来源 · 正文与样式均内嵌，可离线阅读；外部资料链接需联网访问。</div>
</main>
</div>
</body>
</html>
'''
output.parent.mkdir(parents=True, exist_ok=True)
output.write_text(page, encoding='utf-8')

# Check that conversion preserves every curated source and local navigation target.
result = BeautifulSoup(page, 'html.parser')
links = {a['href'] for a in result.select('a[href]')}
sources = json.loads((root/'research'/'claude-ban-fingerprint-sources.json').read_text(encoding='utf-8'))
assert all(item['url'] in links for item in sources)
ids = {tag['id'] for tag in result.select('[id]')}
assert all(href[1:] in ids for href in links if href.startswith('#'))
assert len(result.select('table')) == len(soup.select('table'))
print(f'Created: {output}\nVerified: {len(sources)} source links; {len(navigation)} sections; {len(result.select("table"))} tables.')
