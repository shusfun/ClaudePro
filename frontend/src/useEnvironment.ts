import {useEffect, useRef, useState} from 'react';
import {Dialogs, Events} from '@wailsio/runtime';
import * as API from '../bindings/claudepro/service';
import type {State, LocalReport, NetworkReport} from '../bindings/claudepro/models';
import type {Catalog, Config, Clients, Choice, Conflict, Preset, Target} from '../bindings/claudepro/internal/core/models';
import {useToast} from '@claudepro/ui';

export const fields = ['timezone', 'culture', 'region'] as const;
export const labels = {timezone: '系统时区', culture: '区域格式', region: '国家 / 地区'};
const states: Record<string,string> = {active:'目标环境已启用', restored:'原环境已恢复', switching:'正在切换', restoring:'正在恢复', needs_restore:'需要恢复'};
const blank: Config = {version:1,target:{timezone:'',culture:'',region:''},cliPath:'',desktopPath:'',workDir:''};
function errorText(e: unknown) { return e instanceof Error ? e.message : String(e); }

export type Tab = 'home'|'settings'|'history'|'diagnostics';
export function useEnvironment() {
 const toast=useToast();
 const operation=useRef(false);
 const [state,setState]=useState<State>();
 const [catalog,setCatalog]=useState<Catalog>();
 const [clients,setClients]=useState<Clients>();
 const [presets,setPresets]=useState<Preset[]>([]);
 const [localReport,setLocalReport]=useState<LocalReport>();
 const [localError,setLocalError]=useState('');
 const [localBusy,setLocalBusy]=useState(false);
 const [networkReport,setNetworkReport]=useState<NetworkReport>();
 const [networkBusy,setNetworkBusy]=useState(false);
 const [networkError,setNetworkError]=useState('');
 const [networkClock,setNetworkClock]=useState(Date.now());
 const networkCooldown=networkReport?.retryAt?Math.max(0,Math.ceil((Date.parse(networkReport.retryAt)-networkClock)/1000)):0;
 useEffect(()=>{
  if(!networkReport?.retryAt)return;
  setNetworkClock(Date.now());const timer=window.setInterval(()=>setNetworkClock(Date.now()),1000);
  return()=>window.clearInterval(timer);
 },[networkReport?.retryAt]);
 const networkPending=useRef(false);
 const localPending=useRef<Promise<void>|null>(null);
 const [draft,setDraft]=useState<Config>(blank);
 const [tab,setTab]=useState<'home'|'settings'|'history'|'diagnostics'>('home');
 const [busy,setBusy]=useState(false);
 const [loading,setLoading]=useState(true);
 const [error,setError]=useState('');
 const [progress,setProgress]=useState('');
 const [exit,setExit]=useState(false);
 const [conflicts,setConflicts]=useState<Conflict[]>([]);
 const [decisions,setDecisions]=useState<Record<string,string>>({});
 const journal=state?.view.journal;
 const active=!!journal && journal.status!=='restored';
 const configured=!!draft.target.timezone && !!draft.target.culture && !!draft.target.region;
 const disabled=busy||loading;
 const status=busy ? progress||'正在处理…' : journal ? states[journal.status]||journal.status : '日常环境';
 const label=(field:typeof fields[number],value?:string)=>{
  if(!value)return '尚未设置';
  const list=field==='timezone'?catalog?.timezones:field==='culture'?catalog?.cultures:catalog?.regions;
  return list?.find(o=>o.id===value)?.name||value;
 };
 const checkLocal=(target:Target=draft.target)=>{
  if(localPending.current)return localPending.current;
  setLocalBusy(true);setLocalError('');
  const job=API.GetDiagnostics(target).then(async report=>{
   setLocalReport(report);
   try{setState(await API.GetState());}catch(e){setState(undefined);setError(errorText(e));}
  }).catch(e=>setLocalError(errorText(e))).finally(()=>{setLocalBusy(false);localPending.current=null;});
  localPending.current=job;return job;
 };
 const checkNetwork=async()=>{
  if(networkPending.current||(networkReport?.retryAt&&Date.parse(networkReport.retryAt)>Date.now()))return;
  networkPending.current=true;setNetworkBusy(true);setNetworkError('');setNetworkReport(undefined);
  try{setNetworkReport(await API.DetectNetwork());}catch(e){setNetworkError(errorText(e));}
  finally{networkPending.current=false;setNetworkBusy(false);}
 };
 const refresh=async(resetDraft=false)=>{
  const data=await API.GetState();setState(data);
  if(resetDraft)setDraft(data.config);
  return data;
 };
 useEffect(()=>{
  let cancelled=false;
  const init=async()=>{
   try {
    const data=await API.GetState();if(cancelled)return;
    setState(data);setDraft(data.config);
    const results=await Promise.allSettled([API.GetCatalog(),API.GetClients(),API.GetPresets()]);if(cancelled)return;
    if(results[0].status==='fulfilled')setCatalog(results[0].value);else setError(errorText(results[0].reason));
    if(results[1].status==='fulfilled')setClients(results[1].value);else setError(errorText(results[1].reason));
    if(results[2].status==='fulfilled')setPresets(results[2].value||[]);else setError(errorText(results[2].reason));
    await checkLocal(data.config.target);
   }catch(e){if(!cancelled){setError(errorText(e));await checkLocal(blank.target);}}finally{if(!cancelled)setLoading(false);}
  };void init();
  const p=Events.On('operation-progress',(e:{data:unknown})=>setProgress(String(e.data)));
  const q=Events.On('close-requested',()=>setExit(true));
  const b=Events.On('close-blocked',(e:{data:unknown})=>setError(String(e.data)));
  return()=>{cancelled=true;p();q();b();};
 },[]);
 const act=async(fn:()=>Promise<unknown>,success:string)=>{
  if(disabled||operation.current)return;
  operation.current=true;setBusy(true);setError('');setProgress('正在检查…');
  try{await localPending.current;await fn();if(success)toast(success);}catch(e){setError(errorText(e));}
  finally{try{await refresh();}catch(e){setError(errorText(e));}await checkLocal();setBusy(false);setProgress('');operation.current=false;}
 };
 const restore=()=>act(async()=>{
  const latest=await refresh();const cs=latest.view.conflicts||[];
  if(cs.length){setDecisions({});setConflicts(cs);return;}
  await API.Restore({});
 },'');
 const commitRestore=()=>act(async()=>{
  const choices:Record<string,Choice>={};
  for(const c of conflicts) choices[c.field]={action:decisions[c.field],observed:c.current};
  try{await API.Restore(choices);setConflicts([]);}
  catch(e){const latest=await refresh();setConflicts(latest.view.conflicts||[]);setDecisions({});throw e;}
 },'恢复流程完成');
 const save=()=>act(async()=>{await API.SaveConfig(draft);await refresh(true);setTab('home');},'目标环境已保存');
 const reload=()=>act(async()=>{await refresh();setCatalog(await API.GetCatalog());setClients(await API.GetClients());setPresets(await API.GetPresets()||[]);},'已重新读取本机状态');
 const browse=async(field:'workDir'|'cliPath'|'desktopPath')=>{
  if(disabled||operation.current)return;
  operation.current=true;setBusy(true);setError('');
  try{
   const path=await Dialogs.OpenFile({Title:field==='workDir'?'选择 CLI 工作目录':'选择 Claude 客户端',CanChooseDirectories:field==='workDir',CanChooseFiles:field!=='workDir',AllowsMultipleSelection:false,Filters:field==='workDir'?[]:[{DisplayName:'客户端文件',Pattern:field==='cliPath'?'*.exe;*.cmd;*.ps1':'*.exe'}]});
   if(path)setDraft(d=>({...d,[field]:path}));
  }catch(e){setError(errorText(e));}finally{setBusy(false);operation.current=false;}
 };

 const confirmExit=async()=>{
  if(operation.current)return;
  operation.current=true;setBusy(true);setError('');
  try{await API.ConfirmExit();}catch(e){setError(errorText(e));}
  finally{operation.current=false;setBusy(false);}
 };
 return {presets,localReport,localError,localBusy,checkLocal,networkReport,networkError,networkBusy,networkCooldown,checkNetwork,state,catalog,clients,draft,setDraft,tab,setTab,busy,loading,error,setError,progress,exit,setExit,
  conflicts,setConflicts,decisions,setDecisions,journal,active,configured,disabled,status,label,
  restore,commitRestore,save,reload,browse,confirmExit,
  switchEnvironment:()=>act(()=>API.SwitchTo(draft.target),'切换完成，可以启动 Claude'),
  launch:(client:'cli'|'desktop')=>act(()=>API.Launch(client),client==='cli'?'CLI 启动请求已发送':'Desktop 启动请求已发送')};
}
export type EnvironmentModel = ReturnType<typeof useEnvironment>;
