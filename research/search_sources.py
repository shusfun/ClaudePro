import requests, bs4, concurrent.futures, json, pathlib, sys, base64
sys.stdout.reconfigure(encoding='utf-8')
OUT=pathlib.Path(__file__).parent
queries=sys.argv[1:] or ['"Claude" "fingerprint" "ban"','"Claude" "封号" "指纹"','"Claude Code" "fingerprint"','site:github.com claude fingerprint ban','site:anthropic.com account banned privacy device']
def run(q):
    try:
        r=requests.get('https://www.bing.com/search',params={'q':q,'count':10,'setlang':'en-us'},timeout=25)
        soup=bs4.BeautifulSoup(r.text,'html.parser')
        results=[]
        for li in soup.select('li.b_algo'):
            a=li.select_one('h2 a')
            if not a: continue
            u=a.get('href','')
            if 'bing.com/ck/a?' in u:
                from urllib.parse import urlparse,parse_qs
                code=parse_qs(urlparse(u).query).get('u',[''])[0]
                if code.startswith('a1'):
                    try:u=base64.urlsafe_b64decode(code[2:]+'='*((-len(code[2:]))%4)).decode()
                    except Exception:pass
            results.append({'title':a.get_text(' ',strip=True),'url':u,'snippet':li.get_text(' ',strip=True)})
        return {'q':q,'status':r.status_code,'results':results,'fallback':soup.get_text(' ',strip=True)[:1500] if not results else ''}
    except Exception as e:return {'q':q,'error':str(e)}
results=list(concurrent.futures.ThreadPoolExecutor(max_workers=6).map(run,queries))
with (OUT/'search_log.jsonl').open('a',encoding='utf8') as f:
    for x in results:f.write(json.dumps(x,ensure_ascii=False)+'\n')
print(json.dumps(results,ensure_ascii=False,indent=2))
