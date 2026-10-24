import requests,concurrent.futures,json,pathlib,sys
sys.stdout.reconfigure(encoding='utf8')
kind=sys.argv[1]
queries=sys.argv[2:]
def run(q):
    try:
        r=requests.get('https://api.github.com/search/'+kind,params={'q':q,'per_page':20},timeout=30)
        d=r.json()
        if kind=='repositories':
            items=[{k:x.get(k) for k in ['full_name','html_url','description','stargazers_count','pushed_at','archived']} for x in d.get('items',[])]
        else:items=[{k:x.get(k) for k in ['title','html_url','state','created_at','body']} for x in d.get('items',[])]
        return {'query':q,'status':r.status_code,'total':d.get('total_count'),'items':items,'error':d.get('message')}
    except Exception as e:return {'query':q,'error':str(e)}
data=list(concurrent.futures.ThreadPoolExecutor(4).map(run,queries))
with (pathlib.Path(__file__).parent/'github_log.jsonl').open('a',encoding='utf8') as f:
    for x in data:f.write(json.dumps(x,ensure_ascii=False)+'\n')
for x in data:
    for item in x.get('items',[]):
        if item.get('body'):item['body']=item['body'][:2500]
print(json.dumps(data,ensure_ascii=False,indent=2))
