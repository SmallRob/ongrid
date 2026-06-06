import { Fragment, useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { cn } from '@/lib/cn';
import { useI18n } from '@/i18n/locale';
import type {
  ActionsSummary,
  ChangeFact,
  FleetFacts,
  HeroStat,
  KeyIncident,
  Paragraph,
  ReportContent as ReportContentT,
  ResourceFacts,
} from '@/api/reports';

// ReportContent renders a ContentJSON report body — the rich in-app
// view (HLD-014 §前端渲染). Zero chart deps: count-up via rAF, sparkline
// as inline SVG, entity chips via token parsing.

// --- count-up hook (rAF, no deps) ---
function useCountUp(target: number, durationMs = 800): number {
  const [val, setVal] = useState(0);
  const startRef = useRef<number | null>(null);
  useEffect(() => {
    startRef.current = null;
    let raf = 0;
    const step = (ts: number) => {
      if (startRef.current === null) startRef.current = ts;
      const p = Math.min(1, (ts - startRef.current) / durationMs);
      // easeOutCubic
      const eased = 1 - Math.pow(1 - p, 3);
      setVal(target * eased);
      if (p < 1) raf = requestAnimationFrame(step);
      else setVal(target);
    };
    raf = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf);
  }, [target, durationMs]);
  return val;
}

function fmtNum(v: number): string {
  return Number.isInteger(v) ? String(v) : v.toFixed(1);
}

// --- inline SVG sparkline ---
function Sparkline({ points, className }: { points: number[]; className?: string }) {
  if (!points || points.length < 2) return null;
  const w = 60;
  const h = 16;
  const max = Math.max(...points, 1);
  const min = Math.min(...points, 0);
  const span = max - min || 1;
  const step = w / (points.length - 1);
  const d = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'}${(i * step).toFixed(1)},${(h - ((p - min) / span) * h).toFixed(1)}`)
    .join(' ');
  return (
    <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} className={className} aria-hidden="true">
      <path d={d} fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function HeroCard({ stat }: { stat: HeroStat }) {
  const animated = useCountUp(stat.value);
  const delta = stat.delta_pct;
  // Lower-is-better metrics (incidents/mttr) get green on a drop; we keep
  // it simple — down = green, up = red. Neutral when ~0.
  const deltaColor =
    delta === undefined ? '' : delta < 0 ? 'text-emerald-400' : delta > 0 ? 'text-red-400' : 'text-zinc-500';
  const arrow = delta === undefined ? '' : delta < 0 ? '↓' : delta > 0 ? '↑' : '→';
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/50 p-3">
      <div className="flex items-baseline gap-1">
        <span className="text-2xl font-semibold tabular-nums text-zinc-100">{fmtNum(animated)}</span>
        {stat.unit && <span className="text-xs text-zinc-500">{stat.unit}</span>}
      </div>
      <div className="mt-0.5 text-[11px] uppercase tracking-wide text-zinc-500">{stat.label}</div>
      <div className="mt-1.5 flex items-center justify-between">
        {stat.sparkline && stat.sparkline.length >= 2 ? (
          <Sparkline points={stat.sparkline} className="text-indigo-400" />
        ) : (
          <span />
        )}
        {delta !== undefined && (
          <span className={cn('text-[11px] font-medium tabular-nums', deltaColor)}>
            {arrow}
            {Math.abs(delta).toFixed(0)}%
          </span>
        )}
      </div>
    </div>
  );
}

// EntityText parses {{entity:kind:id|name}} tokens into clickable chips.
const ENTITY_RE = /\{\{entity:([a-z]+):(\d+)\|([^}]*)\}\}/g;

function entityHref(kind: string, id: string): string | null {
  switch (kind) {
    case 'edge':
      return `/devices/${id}`;
    case 'incident':
      return `/alerts/incidents/${id}`;
    default:
      return null;
  }
}

function EntityText({ text }: { text: string }) {
  const parts: React.ReactNode[] = [];
  let last = 0;
  let m: RegExpExecArray | null;
  ENTITY_RE.lastIndex = 0;
  let i = 0;
  while ((m = ENTITY_RE.exec(text)) !== null) {
    if (m.index > last) parts.push(<Fragment key={`t${i}`}>{text.slice(last, m.index)}</Fragment>);
    const [, kind, id, name] = m;
    const href = entityHref(kind, id);
    parts.push(
      href ? (
        <Link
          key={`e${i}`}
          to={href}
          className="mx-0.5 inline-flex items-center rounded border border-indigo-500/40 bg-indigo-500/10 px-1 py-0.5 text-[12px] text-indigo-300 hover:bg-indigo-500/20"
        >
          {name}
        </Link>
      ) : (
        <span key={`e${i}`} className="mx-0.5 rounded bg-zinc-800 px-1 text-[12px] text-zinc-300">
          {name}
        </span>
      ),
    );
    last = m.index + m[0].length;
    i++;
  }
  if (last < text.length) parts.push(<Fragment key="tail">{text.slice(last)}</Fragment>);
  return <>{parts}</>;
}

const SEV_DOT: Record<string, string> = {
  critical: 'bg-red-500',
  warning: 'bg-amber-500',
  info: 'bg-sky-500',
};

function IncidentRow({ ki }: { ki: KeyIncident }) {
  return (
    <Link
      to={`/alerts/incidents/${ki.id}`}
      className="flex items-center gap-2 rounded-md border border-zinc-800 bg-zinc-900/40 px-3 py-2 text-sm hover:border-zinc-700"
    >
      <span className={cn('h-2 w-2 shrink-0 rounded-full', SEV_DOT[ki.severity] ?? 'bg-zinc-600')} />
      <span className="text-zinc-400">I-{ki.id}</span>
      <span className="flex-1 truncate text-zinc-200">{ki.title}</span>
      {ki.root_cause_snippet && (
        <span className="hidden truncate text-xs text-zinc-500 md:inline">{ki.root_cause_snippet}</span>
      )}
      <span className="shrink-0 text-xs text-zinc-500">
        {ki.duration_min}m · {ki.status}
      </span>
    </Link>
  );
}

function ActionsPanel({ a }: { a: ActionsSummary }) {
  const { tr } = useI18n();
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/40 p-3 text-sm text-zinc-300">
      <div>
        {tr('变更动作', 'Mutating')}: <span className="font-medium text-zinc-100">{a.mutating_total}</span>
        {a.mutating_total > 0 && (
          <span className="text-zinc-500">
            {' '}
            ({tr('已批准', 'approved')} {a.mutating_approved})
          </span>
        )}
        {' · '}
        {tr('只读诊断', 'Read-only')}: <span className="font-medium text-zinc-100">{a.safe_total}</span>
      </div>
      {a.by_tool && a.by_tool.length > 0 && (
        <div className="mt-1.5 flex flex-wrap gap-1.5">
          {a.by_tool.map((t) => (
            <span key={t.tool} className="rounded bg-zinc-800 px-1.5 py-0.5 text-xs text-zinc-400">
              {t.tool} ×{t.count}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

export function ReportContentView({ content }: { content: ReportContentT }) {
  const { tr } = useI18n();
  const paras: Paragraph[] = content.narrative?.paragraphs ?? [];
  const incidents = content.key_incidents ?? [];
  const advice = content.advice ?? [];
  const changes = content.changes ?? [];
  const actions = content.actions_summary;
  const hasActivity =
    incidents.length > 0 ||
    (actions && (actions.mutating_total > 0 || actions.safe_total > 0));

  return (
    <div className="space-y-7">
      {/* Hero stats */}
      {content.hero && content.hero.length > 0 && (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          {content.hero.map((h) => (
            <HeroCard key={h.key} stat={h} />
          ))}
        </div>
      )}

      {/* Narrative */}
      {content.narrative?.headline && (
        <section>
          <h2 className="mb-2 text-base font-semibold text-zinc-100">{content.narrative.headline}</h2>
          {paras.length > 0 && (
            <div className="space-y-2 text-sm leading-relaxed text-zinc-300">
              {paras.map((p, i) => (
                <p key={i}>
                  <EntityText text={p.text} />
                </p>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Resource trend (fleet avg/peak over the period) */}
      {content.resource?.available && (
        <Section title={tr('资源使用（周期 均值 / 峰值）', 'Resource usage (period avg / peak)')}>
          <ResourcePanel r={content.resource} />
        </Section>
      )}

      {/* Monitoring coverage */}
      {content.fleet && content.fleet.total > 0 && (
        <Section title={tr('监控覆盖', 'Monitoring coverage')}>
          <FleetPanel f={content.fleet} />
        </Section>
      )}

      {/* Changes this period */}
      {changes.length > 0 && (
        <Section title={tr('变更记录', 'Changes')}>
          <div className="space-y-1">
            {changes.map((ch, i) => (
              <ChangeRow key={i} c={ch} />
            ))}
          </div>
        </Section>
      )}

      {/* Alerts & response — secondary, only when there's activity */}
      <Section title={tr('告警与处置', 'Alerts & response')}>
        {hasActivity ? (
          <div className="space-y-1.5">
            {incidents.map((ki) => (
              <IncidentRow key={ki.id} ki={ki} />
            ))}
            {actions && (actions.mutating_total > 0 || actions.safe_total > 0) && (
              <ActionsPanel a={actions} />
            )}
          </div>
        ) : (
          <EmptyRow text={tr('本周期无告警，agent 无介入', 'No alerts, no agent intervention this period')} />
        )}
      </Section>

      {/* Recommendations */}
      {advice.length > 0 && (
        <Section title={tr('建议', 'Recommendations')}>
          <ul className="space-y-1.5 text-sm text-zinc-300">
            {advice.map((a, i) => (
              <li key={i} className="flex gap-2">
                <span className="text-indigo-400">•</span>
                <span>
                  <EntityText text={a.text} />
                </span>
              </li>
            ))}
          </ul>
        </Section>
      )}
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h3 className="mb-2 text-xs font-medium uppercase tracking-wide text-zinc-500">{title}</h3>
      {children}
    </section>
  );
}

function ResourcePanel({ r }: { r: ResourceFacts }) {
  const { tr } = useI18n();
  const rows: { label: string; avg: number; peak: number }[] = [
    { label: 'CPU', avg: r.cpu_avg, peak: r.cpu_peak },
    { label: tr('内存', 'Memory'), avg: r.mem_avg, peak: r.mem_peak },
    { label: tr('磁盘', 'Disk'), avg: r.disk_avg, peak: r.disk_peak },
  ];
  return (
    <div className="grid grid-cols-1 gap-2 sm:grid-cols-3">
      {rows.map((row) => (
        <div key={row.label} className="rounded-lg border border-zinc-800 bg-zinc-900/40 p-3">
          <div className="text-[11px] uppercase tracking-wide text-zinc-500">{row.label}</div>
          <div className="mt-1 flex items-baseline gap-3">
            <span>
              <span className="text-lg font-semibold tabular-nums text-zinc-100">{row.avg.toFixed(1)}</span>
              <span className="text-xs text-zinc-500">% {tr('均', 'avg')}</span>
            </span>
            <span className="text-zinc-700">·</span>
            <span>
              <span className="text-sm tabular-nums text-zinc-300">{row.peak.toFixed(1)}</span>
              <span className="text-xs text-zinc-500">% {tr('峰', 'peak')}</span>
            </span>
          </div>
          {/* thin utilisation bar (peak) */}
          <div className="mt-2 h-1 overflow-hidden rounded-full bg-zinc-800">
            <div
              className={cn(
                'h-full rounded-full',
                row.peak >= 85 ? 'bg-red-500' : row.peak >= 60 ? 'bg-amber-500' : 'bg-indigo-500',
              )}
              style={{ width: `${Math.min(100, Math.max(2, row.peak))}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}

function FleetPanel({ f }: { f: FleetFacts }) {
  const { tr } = useI18n();
  const roles = f.roles ?? {};
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900/40 p-3 text-sm text-zinc-300">
      <div>
        {tr('监控设备', 'Devices')}: <span className="font-medium text-zinc-100">{f.total}</span>
        {' · '}
        {tr('在线', 'Online')}: <span className="font-medium text-zinc-100">{f.online}</span>
        {f.total > 0 && <span className="text-zinc-500"> ({Math.round((f.online / f.total) * 100)}%)</span>}
      </div>
      {Object.keys(roles).length > 0 && (
        <div className="mt-1.5 flex flex-wrap gap-1.5">
          {Object.entries(roles).map(([role, n]) => (
            <span key={role} className="rounded bg-zinc-800 px-1.5 py-0.5 text-xs text-zinc-400">
              {role} ×{n}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

function ChangeRow({ c }: { c: ChangeFact }) {
  return (
    <div className="flex items-center gap-2 rounded-md border border-zinc-800 bg-zinc-900/30 px-3 py-1.5 text-xs text-zinc-400">
      <span className="tabular-nums text-zinc-500">{c.at.slice(5, 16).replace('T', ' ')}</span>
      <span className="rounded bg-zinc-800 px-1.5 py-0.5 text-zinc-300">{c.action}</span>
      {c.resource_name && <span className="truncate text-zinc-400">{c.resource_name}</span>}
      {c.actor && <span className="ml-auto shrink-0 text-zinc-600">{c.actor}</span>}
    </div>
  );
}

function EmptyRow({ text }: { text: string }) {
  return (
    <div className="rounded-md border border-dashed border-zinc-800 bg-zinc-900/20 px-3 py-3 text-sm text-zinc-500">
      {text}
    </div>
  );
}
