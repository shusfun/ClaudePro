import {History as HistoryIcon} from 'lucide-react';
import {Card, CardContent, PageHeader, PageState, SectionHeader, StatusLamp} from '@claudepro/ui';
import {fields, labels, type EnvironmentModel} from '../useEnvironment';

export function History({m}:{m:EnvironmentModel}) {
  const journal=m.journal;
  return <div className="page-stack">
    <PageHeader eyebrow="CLAUDEPRO / HISTORY" title="恢复记录" description="原始值、执行步骤与恢复结果都保存在本机。"/>
    <Card><CardContent className="history-content"><SectionHeader title="当前 / 最近一次操作" description={journal?.id||'尚无操作记录'}/>
      {journal?<><div className="originals">{fields.map(f=><div key={f}><small>{labels[f]} · 切换前</small><strong>{m.label(f,journal.original?.[f]?.label)}</strong></div>)}</div>
        <ol className="timeline">{[...(journal.log||[])].reverse().map((entry,i)=><li key={i}><StatusLamp tone={i===0?'active':'off'} size="sm"/><div><strong>{entry.message}</strong><time>{entry.time} · UTC</time></div></li>)}</ol></>
        :<PageState kind="empty" title="尚无恢复记录" detail="首次切换后，这里会显示备份与恢复步骤。"/>}
      <div className="backup-path"><SectionHeader level="h3" size="sm" title={<><HistoryIcon size={15}/>本地备份目录</>}/><code>{m.state?.dataDir}</code><p>journal.json 保存当前恢复记录，history 保存之前的记录。存在未恢复操作时，请勿删除这些文件。</p></div>
    </CardContent></Card>
  </div>;
}
