import {test,expect,type Page} from '@playwright/test';

const original={timezone:{label:'China Standard Time',data:{id:'China Standard Time'}},culture:{label:'zh-CN',data:{LocaleName:'zh-CN',sShortDate:'yyyy~MM~dd'}},region:{label:'45',data:{Nation:'45'}}};
const target={timezone:'Eastern Standard Time',culture:'en-US',region:'244'};
const catalog={timezones:[{id:'China Standard Time',name:'(UTC+08:00) 北京'},{id:target.timezone,name:'(UTC-05:00) 美国东部'}],cultures:[{id:'zh-CN',name:'中文（中国） / zh-CN'},{id:'en-US',name:'英语（美国） / en-US'}],regions:[{id:'45',name:'中国 (CN)'},{id:'244',name:'美国 (US)'}]};

async function mockHost(page:Page,configured=true,longCatalog=false){
 const jsErrors:string[]=[];page.on('pageerror',e=>jsErrors.push(e.message));
 const state:any={view:{current:structuredClone(original),journal:null,conflicts:[]},config:{version:1,target:configured?{...target}:{timezone:'',culture:'',region:''},cliPath:'',desktopPath:'',workDir:'C:\\开发项目'},dataDir:'C:\\Users\\测试用户\\AppData\\Local\\ClaudePro',demo:false,warning:'',appVersion:'0.1.0',allowUpdateInstall:true};
 const ctl={networkData:{country:'日本',countryCode:'JP',city:'Tokyo',region:'Tokyo',timezone:'Asia/Tokyo',recommendedPreset:'japan',recommendationReason:'出口国家为日本，推荐日本模板。'},switchCalls:0,networkCalls:0,networkFail:false,holdNetwork:null as Promise<void>|null,failSwitch:false,failRead:false,restoreCalls:0,launchCalls:[] as string[],holdRestore:null as Promise<void>|null,choices:null as any,update:null as any,skip:'',installCalls:0};
 const options=structuredClone(catalog);
 options.timezones.push({id:'Pacific Standard Time',name:'(UTC-08:00) 美国西部'},{id:'W. Europe Standard Time',name:'欧洲中部'},{id:'Singapore Standard Time',name:'新加坡'},{id:'Tokyo Standard Time',name:'(UTC+09:00) 大阪、东京'});options.cultures.push({id:'de-DE',name:'德语 / de-DE'},{id:'en-SG',name:'英语 / en-SG'},{id:'ja-JP',name:'日语（日本） / ja-JP'});options.regions.push({id:'94',name:'德国 (DE)'},{id:'215',name:'新加坡 (SG)'},{id:'122',name:'日本 (JP)'});
 if(longCatalog)options.cultures.push(...Array.from({length:603},(_,i)=>({id:'test-'+i,name:'测试区域格式 '+String(i).padStart(3,'0')})));
 await page.route('**/wails/runtime',async route=>{
  const req=route.request().postDataJSON();let data:any=null;
  if(req.object!==0){await route.fulfill({json:null});return;}
  const {methodID,args}=req.args;
  const fail=async(message:string)=>route.fulfill({status:500,json:{kind:'RuntimeError',message}});
  switch(methodID){
   case 1244899848:if(ctl.failRead){await fail('备份损坏，修改功能已停用');return;}data=state;break;
   case 2375012986:data=options;break;
   case 889869593:data=[
    {id:'us-east',name:'美国东部',country:'US',target,unavailable:''},
    {id:'us-west',name:'美国西部',country:'US',target:{timezone:'Pacific Standard Time',culture:'en-US',region:'244'},unavailable:''},
    {id:'germany',name:'德国',country:'DE',target:{timezone:'W. Europe Standard Time',culture:'de-DE',region:'94'},unavailable:''},
    {id:'singapore',name:'新加坡',country:'SG',target:{timezone:'Singapore Standard Time',culture:'en-SG',region:'215'},unavailable:''},
    {id:'japan',name:'日本',country:'JP',target:{timezone:'Tokyo Standard Time',culture:'ja-JP',region:'122'},unavailable:''},
    {id:'uk',name:'英国',country:'GB',target:{timezone:'GMT Standard Time',culture:'en-GB',region:'242'},unavailable:'本机系统缺少选项：GMT Standard Time'}];break;
   case 4181845581:data={time:'2026-10-24T05:00:00Z',checks:[
    {id:'timezone',name:'系统时区',status:'info',detail:'当前与目标不同',action:'settings'},
    {id:'cli',name:'Claude Code CLI',status:'ok',detail:'已自动发现客户端',action:'settings'},
    {id:'recovery',name:'恢复记录',status:ctl.failRead?'error':'ok',detail:ctl.failRead?'无法读取恢复记录，修改功能已停用':'没有未恢复操作',action:'history'},
    {id:'proxy-env',name:'代理环境变量',status:'info',detail:'未配置代理环境变量',action:'refresh'}]};break;
   case 33053664:ctl.networkCalls++;if(ctl.holdNetwork)await ctl.holdNetwork;data={time:'2026-10-24T05:01:00Z',source:'https://ipapi.co/json/',status:ctl.networkFail?'unknown':'ok',message:ctl.networkFail?'未能检测：连接失败或请求超时，请重试':'',ip:'203.0.113.8',country:'日本',countryCode:'JP',asn:'AS64496',organization:'Example',proxyMode:'直连（Go 环境变量规则）',...ctl.networkData};break;
   case 4080747073:data={cli:'C:\\Users\\测试用户\\npm\\claude.cmd',desktop:'Claude_pzs8sxrjxfjjc!Claude',desktopKind:'msix'};break;
   case 3855370806:state.config=args[0];break;
   case 2558918817:
   case 3696574784:
    ctl.switchCalls++;if(args[0])state.config.target={...args[0]};
    if(ctl.failSwitch){await fail('请保存工作并关闭 Claude 后重试：claude.exe (PID 123)');return;}
    state.view.current=Object.fromEntries(Object.entries(state.config.target).map(([k,v])=>[k,{label:v,data:{value:v}}]));
    state.view.journal={version:1,id:'20261024T000000',status:'active',original:structuredClone(original),target:state.config.target,steps:[],log:[{time:'2026-10-24T00:00:00Z',message:'切换完成，可以启动 Claude'}]};break;
   case 2575713665:
    ctl.restoreCalls++;ctl.choices=args[0];if(ctl.holdRestore)await ctl.holdRestore;
    state.view.current=structuredClone(original);if(args[0]?.culture?.action==='keep')state.view.current.culture=state.view.conflicts[0].current;state.view.conflicts=[];state.view.journal.status='restored';break;
   case 1184620374:ctl.launchCalls.push(args[0]);if(args[0]==='cli'){await fail('未找到 Claude CLI，请在设置中指定路径');return;}break;
   case 1914844742:data=ctl.update;break;
   case 818344625:ctl.installCalls++;break;
   case 1152682603:ctl.skip=args[0];break;
  }
  await route.fulfill({json:data});
 });
 return {state,ctl,jsErrors};
}
async function select(page:Page,label:string,option:string){
 await page.getByRole('combobox',{name:label,exact:true}).click();
 await page.getByRole('option',{name:option,exact:true}).click();
}
function injectConflict(state:any){
 const changed={label:'fr-FR',data:{LocaleName:'fr-FR',sShortDate:'dd/MM/yyyy'}};
 state.view.current.culture=changed;state.view.conflicts=[{field:'culture',current:changed,original:original.culture,expected:{label:'en-US',data:{value:'en-US'}}}];
}

const status=(page:Page)=>page.locator('.environment-status');
const nav=(page:Page,name:string)=>page.getByRole('navigation',{name:'主导航'}).getByRole('button',{name,exact:true});


test('first setup, switch, conflict review, recovery, clients and exit reminder',async({page})=>{
 const {state,ctl,jsErrors}=await mockHost(page,false);
 await page.goto('/');
 await expect(page.getByRole('heading',{name:'日常环境'})).toBeVisible();
 await expect(page.getByRole('button',{name:'切换环境',exact:true})).toBeDisabled();
 await page.getByRole('button',{name:'美国东部模板'}).click();
 expect(state.config.target.timezone).toBe('');expect(ctl.switchCalls).toBe(0);
 await expect(page.getByRole('heading',{name:'将修改 3 项设置'})).toBeVisible();
 ctl.failSwitch=true;await page.getByRole('button',{name:'切换环境',exact:true}).click();
 await expect(page.getByRole('alert')).toContainText('请保存工作并关闭 Claude');
 ctl.failSwitch=false;await page.getByRole('button',{name:'切换环境',exact:true}).click();
 await expect(status(page)).toContainText('目标环境已启用');
 await expect(page.getByRole('heading',{name:'目标环境生效中'})).toBeVisible();
 await expect(page.getByRole('button',{name:'切换环境',exact:true})).toHaveCount(0);
 await expect(page.getByRole('button',{name:'日本模板'})).toBeDisabled();
 await expect(page.getByRole('button',{name:'自定义区域设置'})).toBeDisabled();
 injectConflict(state);
 const restore=page.getByRole('button',{name:'恢复原环境',exact:true});await restore.click();
 const dialog=page.getByRole('dialog',{name:'发现配置冲突'});
 await expect(dialog).toBeVisible();
 await expect(page.getByRole('button',{name:'确认并继续恢复'})).toBeDisabled();
 await expect(page.getByRole('radio',{name:'保留当前值'})).not.toBeChecked();
 for(let i=0;i<8;i++)await page.keyboard.press('Tab');
 expect(await dialog.evaluate(el=>el.contains(document.activeElement))).toBeTruthy();
 await page.keyboard.press('Escape');await expect(dialog).not.toBeVisible();await expect(restore).toBeFocused();
 await restore.click();await page.getByRole('radio',{name:'保留当前值'}).check();
 await page.screenshot({animations:'disabled',path:'../test-results/ui/conflict-light.png'});
 let release!:()=>void;ctl.holdRestore=new Promise<void>(r=>release=r);
 await page.getByRole('button',{name:'确认并继续恢复'}).click();
 await expect(page.getByRole('button',{name:'确认并继续恢复'})).toBeDisabled();
 await expect(page.getByRole('button',{name:'取消',exact:true})).toBeDisabled();
 await page.keyboard.press('Escape');await page.mouse.click(10,10);await expect(dialog).toBeVisible();
 expect(ctl.restoreCalls).toBe(1);release();
 await expect(dialog).not.toBeVisible();await expect(status(page)).toContainText('原环境已恢复');
 expect(ctl.choices.culture).toEqual({action:'keep',observed:state.view.current.culture});
 await page.getByRole('button',{name:'启动 CLI',exact:true}).click();
 await expect(page.getByRole('alert')).toContainText('未找到 Claude CLI');
 await page.getByRole('button',{name:'启动 Desktop',exact:true}).click();expect(ctl.launchCalls).toEqual(['cli','desktop']);
 await page.evaluate(()=>{(window as any)._wails.dispatchWailsEvent({name:'close-requested',data:true});});
 await expect(page.getByRole('dialog',{name:'环境尚未恢复'})).toBeVisible();
 await expect(page.getByRole('button',{name:'留在应用'})).toBeFocused();
 await page.getByRole('button',{name:'留在应用'}).click();
 expect(jsErrors).toEqual([]);
});

test('custom titlebar, themes, client settings dialog and compact layout',async({page})=>{
 const {state,jsErrors}=await mockHost(page);
 await page.emulateMedia({colorScheme:'light'});await page.goto('/');
 await expect(page.getByRole('heading',{name:'将修改 3 项设置'})).toBeVisible();
 const bar=page.locator('.titlebar');
 await expect(bar).toBeVisible();
 expect(await bar.evaluate(el=>getComputedStyle(el).getPropertyValue('--wails-draggable').trim())).toBe('drag');
 const controls=page.getByRole('group',{name:'窗口控制'});
 for(const name of ['最小化','最大化','关闭'])await expect(controls.getByRole('button',{name})).toBeVisible();
 expect(await controls.getByRole('button',{name:'关闭'}).evaluate(el=>getComputedStyle(el).getPropertyValue('--wails-draggable').trim())).toBe('no-drag');
 await expect(page.locator('html')).toHaveAttribute('data-theme','system');
 await page.screenshot({animations:'disabled',path:'../test-results/ui/overview-light.png'});
 await page.emulateMedia({colorScheme:'dark'});await expect(page.locator('html')).toHaveClass(/dark/);
 await page.screenshot({animations:'disabled',path:'../test-results/ui/overview-dark.png'});
 await page.getByRole('button',{name:'界面主题：跟随系统，点击切换为浅色'}).click();
 await expect(page.locator('html')).not.toHaveClass(/dark/);
 await page.getByRole('button',{name:'界面主题：浅色，点击切换为深色'}).click();
 await page.reload();await expect(page.locator('html')).toHaveAttribute('data-theme','dark');
 await page.emulateMedia({colorScheme:'light'});await expect(page.locator('html')).toHaveClass(/dark/);
 await page.getByRole('button',{name:'客户端设置'}).click();
 const settings=page.getByRole('dialog',{name:'客户端设置',exact:true});await expect(settings).toBeVisible();
 await expect(settings.getByLabel('CLI 工作目录')).toHaveValue('C:\\开发项目');
 await settings.getByLabel('CLI 工作目录').fill('C:\\新目录');
 await page.screenshot({animations:'disabled',path:'../test-results/ui/settings-dark.png'});
 await settings.getByRole('button',{name:'保存设置'}).click();
 await expect(settings).not.toBeVisible();
 expect(state.config.workDir).toBe('C:\\新目录');expect(state.config.target.region).toBe('244');
 await page.setViewportSize({width:820,height:600});
 await expect(nav(page,'恢复记录')).toBeVisible();
 await nav(page,'恢复记录').click();
 await expect(page.getByRole('heading',{name:'恢复记录',exact:true})).toBeVisible();
 await page.getByRole('button',{name:'界面主题：深色，点击切换为跟随系统'}).click();
 await expect(page.locator('html')).not.toHaveClass(/dark/);
 await page.screenshot({animations:'disabled',path:'../test-results/ui/compact-history-light.png'});
 await nav(page,'环境概览').click();
 await expect(page.getByRole('button',{name:'切换环境',exact:true})).toBeInViewport();
 await page.screenshot({animations:'disabled',path:'../test-results/ui/compact-overview-light.png'});
 expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBeTruthy();
 expect(jsErrors).toEqual([]);
});

test('macOS layout leaves room for traffic lights and hides window buttons',async({page})=>{
 await mockHost(page);await page.goto('/?platform=darwin');
 await expect(page.locator('html')).toHaveAttribute('data-platform','darwin');
 await expect(page.locator('.traffic-light-space')).toBeVisible();
 await expect(page.getByRole('group',{name:'窗口控制'})).toHaveCount(0);
 await page.screenshot({animations:'disabled',path:'../test-results/ui/overview-mac.png'});
});

test('interrupted recovery and damaged backup retain protections',async({page})=>{
 const {state,ctl}=await mockHost(page);
 state.view.journal={version:1,id:'interrupted',status:'needs_restore',original,target,steps:[],log:[{time:'2026-10-24T00:00:00Z',message:'发现未完成操作，请恢复原环境'}]};
 state.warning='设置文件损坏，但恢复记录仍可读取。';state.demo=true;
 await page.goto('/');
 await expect(page.getByText('模拟模式：',{exact:false})).toBeVisible();
 await expect(status(page)).toContainText('需要恢复');
 await expect(page.getByRole('heading',{name:'需要恢复原环境'})).toBeVisible();
 await expect(page.getByRole('button',{name:'切换环境',exact:true})).toHaveCount(0);
 for(const name of ['启动 CLI','启动 Desktop'])await expect(page.getByRole('button',{name,exact:true})).toBeDisabled();
 await expect(page.getByRole('button',{name:'恢复原环境',exact:true})).toBeEnabled();
 await nav(page,'恢复记录').click();
 await expect(page.getByText('发现未完成操作，请恢复原环境',{exact:true})).toBeVisible();
 await page.screenshot({animations:'disabled',path:'../test-results/ui/recovery-history.png',fullPage:true});
 ctl.failRead=true;await page.reload();
 await expect(page.getByRole('alert').first()).toContainText('备份损坏');
 await expect(page.getByRole('button',{name:'切换环境',exact:true})).toHaveCount(0);
});

test('long option list remains scrollable and keyboard selectable',async({page})=>{
 await mockHost(page,true,true);await page.setViewportSize({width:820,height:600});await page.goto('/');
 await page.getByRole('button',{name:'自定义区域设置'}).click();
 const field=page.getByRole('combobox',{name:'区域格式',exact:true});
 await field.click();const popup=page.getByRole('listbox');await expect(popup).toBeVisible();
 expect(await popup.evaluate(el=>el.scrollHeight>el.clientHeight)).toBeTruthy();
 await field.fill('测试区域格式 602');await page.keyboard.press('ArrowDown');await page.keyboard.press('Enter');
 await expect(field).toHaveValue('测试区域格式 602');
 expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBeTruthy();
});

test('search previews preserve selected values, drafts survive navigation and restored targets can change',async({page})=>{
 const {state,ctl}=await mockHost(page);await page.goto('/');
 await page.getByRole('button',{name:'自定义区域设置'}).click();
 const region=page.getByRole('combobox',{name:'国家 / 地区',exact:true});
 await region.fill('jP');await page.keyboard.press('ArrowDown');await page.keyboard.press('Enter');
 await expect(region).toHaveValue('日本 (JP)');
 const culture=page.getByRole('combobox',{name:'区域格式',exact:true});
 await culture.fill('no-such-option');await expect(page.getByText('没有匹配的选项，请尝试名称或代码')).toBeVisible();
 await page.keyboard.press('Escape');
 await nav(page,'环境检测').click();await nav(page,'环境概览').click();
 await expect(page.getByRole('combobox',{name:'区域格式',exact:true})).toHaveValue('英语（美国） / en-US');
 await expect(page.getByRole('combobox',{name:'国家 / 地区',exact:true})).toHaveValue('日本 (JP)');
 await select(page,'系统时区','(UTC+09:00) 大阪、东京');
 await page.getByRole('combobox',{name:'区域格式',exact:true}).fill('日语');
 await page.getByRole('option',{name:'日语（日本） / ja-JP'}).click();
 await expect(page.getByRole('button',{name:'日本模板'})).toHaveAttribute('aria-pressed','true');
 await expect(page.getByRole('button',{name:'英国模板'})).toBeDisabled();
 expect(ctl.switchCalls).toBe(0);expect(state.config.target.region).toBe('244');
 await page.getByRole('button',{name:'切换环境',exact:true}).click();
 await expect(status(page)).toContainText('目标环境已启用');
 expect(state.config.target.region).toBe('122');
 await page.getByRole('button',{name:'恢复原环境',exact:true}).click();
 await expect(status(page)).toContainText('原环境已恢复');
 await page.getByRole('button',{name:'美国东部模板'}).click();
 await expect(page.getByRole('button',{name:'美国东部模板'})).toHaveAttribute('aria-pressed','true');
});

test('network is manual, retryable and does not block restoration; diagnostics survive corrupt backup',async({page})=>{
 const {state,ctl}=await mockHost(page);await page.goto('/');
 await page.getByRole('button',{name:'切换环境',exact:true}).click();
 await expect(status(page)).toContainText('目标环境已启用');
 await nav(page,'环境检测').click();expect(ctl.networkCalls).toBe(0);
 await expect(page.getByText('已自动发现客户端',{exact:true})).toBeVisible();
 let release!:()=>void;ctl.holdNetwork=new Promise<void>(r=>release=r);
 await page.getByRole('button',{name:'检测网络出口',exact:true}).click();
 await expect(page.getByRole('button',{name:'检测中…',exact:true})).toBeDisabled();
 await nav(page,'环境概览').click();
 await page.getByRole('button',{name:'恢复原环境',exact:true}).click();
 await expect(status(page)).toContainText('原环境已恢复');
 release();await nav(page,'环境检测').click();
 await expect(page.getByText('203.0.113.8',{exact:true})).toBeVisible();
 await expect(page.getByText('本次检测出口国家与所选目标不同',{exact:false})).toBeVisible();
 await page.screenshot({animations:'disabled',path:'../test-results/ui/diagnostics-light.png',fullPage:true});
 ctl.networkFail=true;await page.getByRole('button',{name:'重新检测并推荐',exact:true}).click();
 await expect(page.getByText('连接失败或请求超时',{exact:false})).toBeVisible();
 ctl.networkFail=false;await page.getByRole('button',{name:'重新检测并推荐',exact:true}).click();
 await expect(page.getByText('203.0.113.8',{exact:true})).toBeVisible();
 await page.reload();await nav(page,'环境检测').click();
 await expect(page.getByText('尚未联网检测。结果仅保留在本次应用会话。')).toBeVisible();expect(ctl.networkCalls).toBe(3);
 ctl.failRead=true;await page.reload();await nav(page,'环境检测').click();
 await expect(page.getByText('无法读取恢复记录，修改功能已停用',{exact:true})).toBeVisible();
 expect(state.view.journal.status).toBe('restored');
});

test('exit timezone recommends a coast, adoption only previews, uncertain locations never guess',async({page})=>{
 const {state,ctl}=await mockHost(page,false);
 ctl.networkData={country:'United States',countryCode:'US',city:'Los Angeles',region:'California',timezone:'America/Los_Angeles',recommendedPreset:'us-west',recommendationReason:'出口时区为 America/Los_Angeles（美国太平洋时间），推荐美国西部模板。'};
 await page.goto('/');await expect(page.getByRole('button',{name:'检测并推荐地区'})).toBeEnabled();expect(ctl.networkCalls).toBe(0);
 await page.getByRole('button',{name:'检测并推荐地区'}).click();
 await expect(page.getByText('推荐：美国西部',{exact:true})).toBeVisible();
 await expect(page.getByText('203.0.113.8 · Los Angeles / California / United States · America/Los_Angeles',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'美国西部模板'}).locator('.preset-tag')).toHaveText('推荐');
 expect(state.config.target.timezone).toBe('');expect(ctl.switchCalls).toBe(0);
 await page.getByRole('button',{name:'采用推荐',exact:true}).click();
 await expect(page.getByRole('button',{name:'美国西部模板'})).toHaveAttribute('aria-pressed','true');
 expect(state.config.target.timezone).toBe('');expect(ctl.switchCalls).toBe(0);
 await page.screenshot({animations:'disabled',path:'../test-results/ui/recommendation-west.png',fullPage:true});
 await page.getByRole('button',{name:'切换环境',exact:true}).click();
 await expect(status(page)).toContainText('目标环境已启用');
 await expect(page.getByRole('button',{name:'采用推荐',exact:true})).toHaveCount(0);
 await page.getByRole('button',{name:'恢复原环境',exact:true}).click();
 await expect(status(page)).toContainText('原环境已恢复');
 ctl.networkData={...ctl.networkData,city:'New York',region:'New York',timezone:'America/New_York',recommendedPreset:'us-east',recommendationReason:'出口时区为 America/New_York（美国东部时间），推荐美国东部模板。'};
 await page.getByRole('button',{name:'重新检测并推荐'}).click();
 await expect(page.getByText('推荐：美国东部',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'采用推荐',exact:true})).toBeEnabled();
 ctl.networkData={...ctl.networkData,city:'Chicago',region:'Illinois',timezone:'America/Chicago',recommendedPreset:'',recommendationReason:'出口时区为 America/Chicago，未匹配现有美东 / 美西模板，请使用自定义。'};
 await page.getByRole('button',{name:'重新检测并推荐'}).click();
 await expect(page.getByText('暂不能推荐现有模板',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'采用推荐',exact:true})).toHaveCount(0);
 await expect(page.getByRole('button',{name:'美国西部模板'})).toHaveAttribute('aria-pressed','true');
 ctl.networkFail=true;await page.getByRole('button',{name:'重新检测并推荐'}).click();
 await expect(page.getByText('暂不能推荐现有模板',{exact:true})).toHaveCount(0);
 await expect(page.getByText('连接失败或请求超时',{exact:false})).toBeVisible();
 expect(ctl.switchCalls).toBe(1);
});

test('multi-source fallback shows provenance, cooldown blocks repeated calls, conflict clears recommendation',async({page})=>{
 const {ctl}=await mockHost(page,false);
 const now=Date.now();await page.clock.install({time:now});
 Object.assign(ctl.networkData,{country:'United States',countryCode:'US',city:'New York',region:'New York',timezone:'America/New_York',recommendedPreset:'us-east',recommendationReason:'出口时区为 America/New_York，推荐美国东部模板。',source:'https://get.geojs.io/v1/ip/geo.json',notes:'https://ipapi.co/json/：服务返回 HTTP 429；交叉查询：https://ipwho.is/ → America/New_York',retryAt:new Date(now+60000).toISOString(),cached:false,locationConflict:false});
 await page.goto('/');await page.getByRole('button',{name:'检测并推荐地区'}).click();
 await expect(page.getByText('推荐：美国东部',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:/秒后可刷新/})).toBeDisabled();
 await page.getByText('查看检测来源与状态',{exact:true}).click();
 await expect(page.locator('.network-details')).toContainText('HTTP 429');expect(ctl.networkCalls).toBe(1);
 await page.getByRole('button',{name:'采用推荐',exact:true}).click();
 await expect(page.getByRole('button',{name:'美国东部模板'})).toHaveAttribute('aria-pressed','true');
 await page.clock.fastForward(61000);
 await expect(page.getByRole('button',{name:'重新检测并推荐'})).toBeEnabled();
 Object.assign(ctl.networkData,{recommendedPreset:'',locationConflict:true,recommendationReason:'检测源的出口 IP 或地区判断存在分歧，暂不推荐模板。',retryAt:'',notes:'GeoJS → America/New_York；ipwho.is → America/Los_Angeles'});
 await page.getByRole('button',{name:'重新检测并推荐'}).click();
 await expect(page.getByText('检测来源存在分歧',{exact:true})).toBeVisible();
 await expect(page.getByRole('button',{name:'采用推荐',exact:true})).toHaveCount(0);
 await page.setViewportSize({width:820,height:600});
 await page.screenshot({animations:'disabled',path:'../test-results/ui/network-conflict-compact.png',fullPage:true});
 expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBeTruthy();
});

test('update dialog appears, can be postponed, then skipped',async({page})=>{
 const {ctl}=await mockHost(page);
 ctl.update={version:'0.2.0',name:'ClaudePro',notes:'修复更新检查',filename:'claudepro.exe',size:12000000,allowInstall:true};
 await page.goto('/');
 await expect(page.getByText('发现新版本',{exact:true})).toBeVisible();
 await expect(page.getByText('claudepro.exe',{exact:false})).toBeVisible();
 await page.getByRole('button',{name:'稍后提醒',exact:true}).click();
 await expect(page.getByText('发现新版本',{exact:true})).toHaveCount(0);
 await page.getByRole('button',{name:'客户端设置'}).click();
 await page.getByRole('button',{name:'检查更新',exact:true}).click();
 await expect(page.getByText('发现新版本',{exact:true})).toBeVisible();
 await page.getByRole('button',{name:'跳过此版本',exact:true}).click();
 await expect(page.getByText('发现新版本',{exact:true})).toHaveCount(0);
 expect(ctl.skip).toBe('0.2.0');
});


test('update install capability comes from update response when environment is unavailable',async({page})=>{
 const {ctl}=await mockHost(page);
 ctl.update={version:'0.1.4',name:'ClaudePro',notes:'修复恢复与更新',filename:'claudepro-macos-universal.zip',size:12000000,allowInstall:true};
 ctl.failRead=true;
 await page.goto('/?platform=darwin');
 await expect(page.getByRole('button',{name:'立即更新',exact:true})).toBeVisible();
 await expect(page.getByText(/当前是开发构建/)).toHaveCount(0);
 await page.getByRole('button',{name:'立即更新',exact:true}).click();
 await expect.poll(()=>ctl.installCalls).toBe(1);
});
