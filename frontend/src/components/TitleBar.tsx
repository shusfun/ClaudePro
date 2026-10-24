import {useEffect, useState, type ReactNode} from 'react';
import {Events, Window} from '@wailsio/runtime';
import {Copy, Minus, Square, X} from 'lucide-react';
import {ClaudeMark} from './ClaudeMark';
import {isMac} from '../platform';

// Windows 无边框窗口的自绘控制按钮；macOS 使用系统红绿灯，这里只留出位置。
function WindowControls() {
  const [maximised,setMaximised]=useState(false);
  useEffect(()=>{
    const sync=()=>{void Window.IsMaximised().then(v=>setMaximised(v===true)).catch(()=>{});};
    sync();
    const offs=[Events.Types.Common.WindowMaximise,Events.Types.Common.WindowUnMaximise,Events.Types.Common.WindowRestore,Events.Types.Common.WindowDidResize].map(name=>Events.On(name,sync));
    return()=>offs.forEach(off=>off());
  },[]);
  const call=(fn:()=>Promise<void>)=>()=>{void fn().catch(()=>{});};
  return <div className="window-controls" role="group" aria-label="窗口控制">
    <button type="button" className="window-control" aria-label="最小化" title="最小化" onClick={call(()=>Window.Minimise())}><Minus/></button>
    <button type="button" className="window-control" aria-label={maximised?'向下还原':'最大化'} title={maximised?'向下还原':'最大化'} onClick={call(()=>Window.ToggleMaximise())}>{maximised?<Copy className="restore-icon"/>:<Square/>}</button>
    <button type="button" className="window-control close" aria-label="关闭" title="关闭" onClick={call(()=>Window.Close())}><X/></button>
  </div>;
}

export function TitleBar({status,tools,menu}:{status:ReactNode;tools:ReactNode;menu?:ReactNode}) {
  const mac=isMac();
  return <header className="titlebar" onDoubleClick={e=>{
    // macOS 隐藏标题栏后双击不再触发缩放，按系统习惯补上；Windows 由原生标题栏命中处理。
    if(mac&&e.target===e.currentTarget)void Window.ToggleMaximise().catch(()=>{});
  }}>
    {mac&&<span className="traffic-light-space" aria-hidden="true"/>}
    {menu}
    <div className="titlebar-brand"><ClaudeMark size={18}/><strong>ClaudePro</strong></div>
    <div className="titlebar-center">{status}</div>
    <div className="titlebar-tools">{tools}</div>
    {!mac&&<WindowControls/>}
  </header>;
}
