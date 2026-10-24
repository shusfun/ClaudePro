import {useEffect, useRef, useState} from 'react';
import {Browser, Events, Updater} from '@wailsio/runtime';
import * as API from '../bindings/claudepro/service';
import type {UpdateInfo} from '../bindings/claudepro/models';
import {useToast} from '@claudepro/ui';

export const releasesURL = 'https://github.com/shusfun/ClaudePro/releases';

type Phase = 'idle'|'available'|'downloading'|'ready'|'error';

function eventData<T>(e:{data?:T}|T):T {
  if(e && typeof e==='object' && 'data' in e && (e as {data?:T}).data!==undefined)return (e as {data:T}).data;
  return e as T;
}

export function useUpdate(demo:boolean|undefined, busy:boolean) {
  const toast=useToast();
  const [open,setOpen]=useState(false);
  const [phase,setPhase]=useState<Phase>('idle');
  const [release,setRelease]=useState<UpdateInfo|null>(null);
  const [progress,setProgress]=useState(0);
  const [error,setError]=useState('');
  const [checking,setChecking]=useState(false);
  const auto=useRef(false);

  const show=(info:UpdateInfo)=>{
    setRelease(info);setError('');setProgress(0);setPhase('available');setOpen(true);
  };

  const check=async(manual:boolean)=>{
    if(demo||checking)return;
    if(manual){setChecking(true);setError('');}
    try{
      const info=await API.CheckUpdate();
      if(info?.version)show(info);
      else if(manual)toast('已是最新版本');
    }catch(e){if(manual)setError(e instanceof Error?e.message:String(e));}
    finally{if(manual)setChecking(false);}
  };

  useEffect(()=>{
    if(demo===true||auto.current)return;
    auto.current=true;void check(false);
  },[demo]);

  useEffect(()=>{
    const offs=[
      Events.On(Updater.Events.DownloadProgress,e=>{
        const p=eventData<{written?:number;total?:number}>(e);
        if(p?.total)setProgress(Math.round((p.written||0)*100/p.total));
      }),
      Events.On(Updater.Events.DownloadStarted,()=>{setPhase('downloading');setProgress(0);setOpen(true);}),
      Events.On(Updater.Events.UpdateReady,()=>{setPhase('ready');setProgress(100);setOpen(true);}),
      Events.On(Updater.Events.Error,e=>{
        const info=eventData<{message?:string}>(e);
        setPhase('error');setError(info?.message||'更新失败');setOpen(true);
      }),
    ];
    return()=>offs.forEach(off=>off());
  },[]);

  const act=async(fn:()=>Promise<unknown>,next?:Phase)=>{
    setError('');
    try{if(next)setPhase(next);await fn();}catch(e){setPhase('error');setError(e instanceof Error?e.message:String(e));}
  };

  return {
    open,phase,release,progress,error,checking,allowInstall:release?.allowInstall,
    busyEnv:busy,
    check:()=>check(true),
    close:()=>{setOpen(false);if(phase!=='downloading'&&phase!=='ready')setPhase('idle');},
    install:()=>act(()=>API.InstallUpdate(),'downloading'),
    restart:()=>act(()=>API.RestartUpdate()),
    skip:()=>act(async()=>{if(release?.version)await API.SkipUpdate(release.version);setOpen(false);setPhase('idle');}),
    openReleases:()=>{void Browser.OpenURL(releasesURL).catch(e=>setError(String(e)));},
  };
}
export type UpdateModel = ReturnType<typeof useUpdate>;
