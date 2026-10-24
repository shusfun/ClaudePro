import requests,bs4,concurrent.futures,json,pathlib,sys,re,hashlib
sys.stdout.reconfigure(encoding='utf8')
OUT=pathlib.Path(__file__).parent/'sources'; OUT.mkdir(exist_ok=True)
def fetch(u):
    try:
        r=requests.get(u,timeout=30)
        if 'api.github.com/repos/' in u and u.endswith('/readme'):
            import base64
            d=r.json(); text=base64.b64decode(d.get('content','')).decode('utf8',errors='replace')
            title=d.get('html_url',u)
        else:
            s=bs4.BeautifulSoup(r.text,'html.parser')
            title=s.title.get_text() if s.title else u
            for e in s(['script','style','nav','footer','header']):e.decompose()
            main=s.find('main') or s.find('article') or s
            text=main.get_text('\n',strip=True)
        id=hashlib.sha256(u.encode()).hexdigest()[:12]
        record={'url':u,'final_url':r.url,'status':r.status_code,'title':title,'text':text}
        (OUT/(id+'.json')).write_text(json.dumps(record,ensure_ascii=False,indent=2),encoding='utf8')
        hits=[]
        for m in re.finditer(r'fingerprint|指纹|device.{0,12}id|封号|ban\b|banned|suspend|privacy|JA[34]|TLS|遥测|https?://(?:linux.do|v2ex.com|[^ /]+\.com/blog)',text,re.I):
            start=max(0,m.start()-150); end=min(len(text),m.end()+450)
            if not hits or start>hits[-1][1]:hits.append((start,end,text[start:end]))
        return {'url':u,'status':r.status_code,'title':title,'file':id+'.json','length':len(text),'start':text[:900],'excerpts':[x[2] for x in hits[:8]]}
    except Exception as e:return {'url':u,'error':str(e)}
print(json.dumps(list(concurrent.futures.ThreadPoolExecutor(8).map(fetch,sys.argv[1:])),ensure_ascii=False,indent=2))
