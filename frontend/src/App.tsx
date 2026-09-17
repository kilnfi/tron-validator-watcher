import { useEffect, useRef, useState } from 'react'
import { Activity, Crown, Hash, TrendingUp, TrendingDown, Zap, Clock, AlertTriangle, Shield, Check, X, Users } from 'lucide-react'

interface RecentBlock {
  number: number
  proposer: string
  address: string
  is_ours: boolean
  missed: boolean
  timestamp: number
}

interface ValidatorStatus {
  name: string
  address: string
  rank: number
  proposed: number
  missed: number
  consec_missed: number
  is_active: boolean
  vote_count: number
  votes_margin_to_sr: number
  balance: number
  reward_balance: number
  frozen: number
  next_slot_in_ms: number
}

interface Status {
  epoch: number
  round_progress: number
  round_duration: number
  next_round_in_ms: number
  is_our_slot: boolean
  current_leader: string
  proposed_total: number
  missed_total: number
  consec_missed: number
  recent_blocks: RecentBlock[]
  validators: ValidatorStatus[]
  total_srs: number
}

function formatDuration(ms: number): string {
  const s = Math.floor(ms / 1000)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${sec}s`
  return `${sec}s`
}

function formatTime(ts: number): string {
  const diff = Math.floor((Date.now() - ts) / 1000)
  if (diff < 5) return 'just now'
  if (diff < 60) return `${diff}s ago`
  return `${Math.floor(diff / 60)}m ago`
}

function successRate(proposed: number, missed: number): string {
  const total = proposed + missed
  if (total === 0) return 'N/A'
  return ((proposed / total) * 100).toFixed(1) + '%'
}

function formatTRX(trx: number): string {
  if (trx >= 1_000_000) return (trx / 1_000_000).toFixed(2) + 'M'
  if (trx >= 1_000) return (trx / 1_000).toFixed(1) + 'K'
  return trx.toFixed(1)
}

function formatVotesMargin(votes: number): string {
  const sign = votes >= 0 ? '+' : '-'
  const abs = Math.abs(votes)
  if (abs >= 1_000_000) return `${sign}${(abs / 1_000_000).toFixed(2)}M`
  if (abs >= 1_000) return `${sign}${(abs / 1_000).toFixed(1)}K`
  return `${sign}${abs}`
}

function rankStyle(rank: number): string {
  if (rank <= 9) return 'text-yellow-400 bg-yellow-500/10 border border-yellow-500/30'
  if (rank <= 18) return 'text-blue-400 bg-blue-500/10 border border-blue-500/30'
  return 'text-violet-400 bg-violet-500/10 border border-violet-500/30'
}

function lastMissAgo(blocks: RecentBlock[], address?: string): string | null {
  const b = blocks.find(b => b.is_ours && b.missed && (!address || b.address === address))
  return b ? formatTime(b.timestamp) : null
}

function truncateAddress(addr: string): string {
  if (addr.length <= 16) return addr
  return `${addr.slice(0, 8)}…${addr.slice(-6)}`
}

export default function App() {
  const [status, setStatus] = useState<Status | null>(null)
  const [error, setError] = useState(false)
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)
  const [nextBlockIn, setNextBlockIn] = useState<number | null>(null)
  const prevMissedRef = useRef<Record<string, number>>({})
  const [flashAddresses, setFlashAddresses] = useState<Set<string>>(new Set())
  const [nextSlotTimes, setNextSlotTimes] = useState<Record<string, number>>({})

  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const res = await fetch('/api/status')
        if (!res.ok) throw new Error()
        const data: Status = await res.json()
        setStatus(data)
        setLastUpdate(new Date())
        setError(false)

        const fetchedAt = Date.now()
        const times: Record<string, number> = {}
        for (const v of data.validators ?? []) {
          if (v.next_slot_in_ms > 0) times[v.address] = fetchedAt + v.next_slot_in_ms
        }
        setNextSlotTimes(times)

        // Detect new misses and trigger flash
        const newFlashes = new Set<string>()
        for (const v of data.validators ?? []) {
          const prev = prevMissedRef.current[v.address] ?? 0
          if (v.missed > prev) newFlashes.add(v.address)
          prevMissedRef.current[v.address] = v.missed
        }
        if (newFlashes.size > 0) {
          setFlashAddresses(newFlashes)
          setTimeout(() => setFlashAddresses(new Set()), 1600)
        }
      } catch {
        setError(true)
      }
    }

    fetchStatus()
    const interval = setInterval(fetchStatus, 3000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    const tick = setInterval(() => {
      setNextBlockIn(3000 - (Date.now() % 3000))
    }, 100)
    return () => clearInterval(tick)
  }, [])

  const progressPct = status
    ? Math.round((status.round_progress / status.round_duration) * 100)
    : 0

  return (
    <div className="min-h-screen p-6 max-w-5xl mx-auto">

      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-blue-500/10 border border-blue-500/20 flex items-center justify-center">
            <Zap className="w-4 h-4 text-blue-400" />
          </div>
          <h1 className="text-lg font-semibold text-slate-100">Tron Validator Watcher</h1>
        </div>
        <div className="flex items-center gap-2">
          {error ? (
            <span className="flex items-center gap-1.5 text-xs font-medium text-red-400 bg-red-500/10 border border-red-500/20 px-3 py-1.5 rounded-full">
              <span className="w-1.5 h-1.5 rounded-full bg-red-400" />
              Disconnected
            </span>
          ) : (
            <span className="flex items-center gap-1.5 text-xs font-medium text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-3 py-1.5 rounded-full">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              Live
            </span>
          )}
          {lastUpdate && (
            <span className="text-xs text-slate-600 font-mono">
              {lastUpdate.toLocaleTimeString()}
            </span>
          )}
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-4 gap-4 mb-4">
        <StatCard
          label="Epoch"
          value={status ? `#${status.epoch}` : '—'}
          icon={<Hash className="w-4 h-4 text-slate-500" />}
          title={status ? `Start: ${new Date(status.epoch * 21600 * 1000).toUTCString()}` : undefined}
        />
        <StatCard
          label="Proposed"
          value={status?.proposed_total ?? '—'}
          icon={<TrendingUp className="w-4 h-4 text-emerald-500" />}
          valueClass="text-emerald-400"
        />
        <StatCard
          label="Missed"
          value={status?.missed_total ?? '—'}
          icon={<TrendingDown className="w-4 h-4 text-red-500" />}
          valueClass={status && status.missed_total > 0 ? 'text-red-400' : 'text-slate-300'}
        />
        <StatCard
          label="Success rate"
          value={status ? successRate(status.proposed_total, status.missed_total) : '—'}
          icon={<Activity className="w-4 h-4 text-blue-500" />}
          valueClass="text-blue-400"
        />
      </div>

      {/* Round progress + current slot */}
      <div className="grid grid-cols-3 gap-4 mb-4">

        {/* Round progress */}
        <div className="col-span-2 bg-[#111827] border border-white/5 rounded-xl p-5">
          <div className="flex justify-between items-center mb-3">
            <span className="text-xs font-semibold uppercase tracking-widest text-slate-500">Round progress</span>
            <span className="text-xs font-mono text-slate-500">
              {status ? `${status.round_progress.toLocaleString()} / ${status.round_duration.toLocaleString()} s` : '—'}
            </span>
          </div>
          <div className="w-full h-2 bg-slate-800 rounded-full overflow-hidden mb-3">
            <div
              className="h-full rounded-full bg-gradient-to-r from-blue-500 to-violet-500 transition-all duration-1000"
              style={{ width: `${progressPct}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-slate-600">
            <span>{progressPct}% complete</span>
            <span className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              Next round in {status ? formatDuration(status.next_round_in_ms) : '—'}
            </span>
          </div>
        </div>

        {/* Last producer */}
        <div className={`bg-[#111827] border rounded-xl p-5 flex flex-col justify-between ${
          status?.recent_blocks[0]?.is_ours
            ? 'border-blue-500/30 bg-blue-500/5'
            : 'border-white/5'
        }`}>
          <span className="text-xs font-semibold uppercase tracking-widest text-slate-500 mb-3">Last producer</span>
          {status ? (
            <>
              <div className="flex items-center gap-2 mb-2">
                {status.recent_blocks[0]?.is_ours && <Crown className="w-4 h-4 text-yellow-400" />}
                <span className="font-semibold text-slate-100 truncate">
                  {status.recent_blocks[0]?.proposer ?? '—'}
                </span>
              </div>
              {status.recent_blocks[0]?.is_ours ? (
                <span className="text-xs font-medium text-blue-400 bg-blue-500/10 border border-blue-500/20 px-2 py-1 rounded-md w-fit">
                  Kiln produced
                </span>
              ) : (
                <span className="text-xs text-slate-600">
                  #{status.recent_blocks[0]?.number ?? '—'}
                </span>
              )}
              {status.consec_missed > 0 && (
                <div className="mt-2 flex items-center gap-1 text-xs text-amber-400">
                  <AlertTriangle className="w-3 h-3" />
                  {status.consec_missed} consecutive miss{status.consec_missed > 1 ? 'es' : ''}
                </div>
              )}
            </>
          ) : (
            <span className="text-slate-600 text-sm">Loading...</span>
          )}
        </div>
      </div>

      {/* Validators */}
      <div className="bg-[#111827] border border-white/5 rounded-xl p-5 mb-4">
        <div className="flex items-center gap-2 mb-4">
          <Users className="w-4 h-4 text-slate-500" />
          <span className="text-xs font-semibold uppercase tracking-widest text-slate-500">Validators</span>
          {status?.validators?.length ? (
            <div className="ml-auto flex items-center gap-3">
              {(status.total_srs ?? 0) > 0 && (
                <span className="text-xs text-slate-600">{status.total_srs} active SRs</span>
              )}
              <span className="text-xs text-slate-600 border-l border-white/10 pl-3">{status.validators.length} monitored</span>
            </div>
          ) : null}
        </div>

        {!status || !status.validators || status.validators.length === 0 ? (
          <div className="text-sm text-slate-600 text-center py-6">Waiting for validator data...</div>
        ) : (
          <div className={`grid gap-3 ${status.validators.length === 1 ? 'grid-cols-1' : 'grid-cols-2'}`}>
            {status.validators.map((v) => (
              <div
                key={v.address}
                className={`rounded-lg border p-4 ${
                  flashAddresses.has(v.address)
                    ? 'animate-flash-red border-red-500/30'
                    : v.consec_missed > 0
                      ? 'border-amber-500/20 bg-amber-500/5'
                      : 'border-white/5 bg-slate-800/30'
                }`}
              >
                {/* Top row: rank + name + active badge */}
                <div className="flex items-center gap-2 mb-2">
                  <span className={`text-xs font-mono font-bold px-1.5 py-0.5 rounded ${rankStyle(v.rank)}`}>
                    #{v.rank}
                  </span>
                  <span className="font-semibold text-slate-100 truncate flex-1">{v.name}</span>
                  {v.is_active ? (
                    <span className="flex items-center gap-1 text-xs font-medium text-emerald-400 bg-emerald-500/10 border border-emerald-500/20 px-2 py-0.5 rounded-full flex-shrink-0">
                      <span className="w-1 h-1 rounded-full bg-emerald-400" />
                      Active
                    </span>
                  ) : (
                    <span className="text-xs font-medium text-slate-600 bg-slate-700/30 border border-slate-700/50 px-2 py-0.5 rounded-full flex-shrink-0">
                      Inactive
                    </span>
                  )}
                </div>

                {/* Next block */}
                {nextSlotTimes[v.address] && (
                  <div className="flex items-center justify-between mb-3 px-3 py-2 rounded-lg bg-yellow-500/5 border border-yellow-500/20">
                    <div className="flex items-center gap-2">
                      <Crown className="w-3.5 h-3.5 text-yellow-400" />
                      <span className="text-xs font-medium text-slate-400">Next block</span>
                    </div>
                    <span className="font-mono font-bold text-yellow-300 text-sm">
                      {new Date(nextSlotTimes[v.address]).toLocaleTimeString()}
                    </span>
                  </div>
                )}

                {/* Address */}
                <a
                  href={`https://tronscan.org/#/address/${v.address}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="font-mono text-xs text-slate-600 hover:text-blue-400 transition-colors mb-3 truncate block"
                  title={v.address}
                >
                  {v.address}
                </a>

                {/* Stats row */}
                <div className="flex items-center gap-4">
                  <div className="flex items-center gap-1.5">
                    <Shield className="w-3 h-3 text-emerald-500" />
                    <span className="text-sm font-bold text-emerald-400">{v.proposed}</span>
                    <span className="text-xs text-slate-600">proposed</span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <X className="w-3 h-3 text-red-500" />
                    <span className={`text-sm font-bold ${v.missed > 0 ? 'text-red-400' : 'text-slate-500'}`}>{v.missed}</span>
                    <span className="text-xs text-slate-600">missed</span>
                  </div>
                  <div className="ml-auto flex flex-col items-end gap-0.5">
                    <span className="text-xs font-mono text-slate-500">{successRate(v.proposed, v.missed)}</span>
                    {v.vote_count > 0 && (
                      <span className="text-xs text-slate-600">{(v.vote_count / 1_000_000).toFixed(1)}M votes</span>
                    )}
                  </div>
                </div>

                {/* SR margin + balances */}
                <div className="mt-3 grid grid-cols-2 gap-2 text-xs">
                  <div className="flex items-center justify-between px-2 py-1.5 rounded-lg bg-slate-800/40 border border-white/5">
                    <span className="text-slate-500">SR margin</span>
                    <span className={`font-mono font-bold ${v.votes_margin_to_sr >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
                      {formatVotesMargin(v.votes_margin_to_sr)}
                    </span>
                  </div>
                  <div className="flex items-center justify-between px-2 py-1.5 rounded-lg bg-slate-800/40 border border-white/5">
                    <span className="text-slate-500">Rewards</span>
                    <span className={`font-mono font-bold ${v.reward_balance > 0 ? 'text-yellow-300' : 'text-slate-500'}`}>
                      {formatTRX(v.reward_balance)}
                    </span>
                  </div>
                  <div className="flex items-center justify-between px-2 py-1.5 rounded-lg bg-slate-800/40 border border-white/5">
                    <span className="text-slate-500">Balance</span>
                    <span className="font-mono text-slate-300">{formatTRX(v.balance)}</span>
                  </div>
                  <div className="flex items-center justify-between px-2 py-1.5 rounded-lg bg-slate-800/40 border border-white/5">
                    <span className="text-slate-500">Staked</span>
                    <span className="font-mono text-slate-300">{formatTRX(v.frozen)}</span>
                  </div>
                </div>
                <div className="mt-1.5 text-xs text-slate-600">
                  {v.proposed + v.missed} block{v.proposed + v.missed !== 1 ? 's' : ''} this round
                  {(() => { const ago = lastMissAgo(status?.recent_blocks ?? [], v.address); return ago ? <span className="ml-2 text-red-500/70">· last miss {ago}</span> : null })()}
                </div>

                {/* Consecutive miss warning */}
                {v.consec_missed > 0 && (
                  <div className="mt-2 flex items-center gap-1 text-xs text-amber-400">
                    <AlertTriangle className="w-3 h-3" />
                    {v.consec_missed} consecutive miss{v.consec_missed > 1 ? 'es' : ''}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Activity feed */}
      <div className="bg-[#111827] border border-white/5 rounded-xl p-5">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Activity className="w-4 h-4 text-slate-500" />
            <span className="text-xs font-semibold uppercase tracking-widest text-slate-500">Recent activity</span>
          </div>
          {nextBlockIn !== null && (
            <div className="flex items-center gap-1.5 text-xs font-mono text-slate-400">
              <Clock className="w-3 h-3 text-slate-500" />
              Next block in{' '}
              <span className={`font-semibold tabular-nums ${nextBlockIn < 500 ? 'text-emerald-400' : 'text-slate-300'}`}>
                {(nextBlockIn / 1000).toFixed(1)}s
              </span>
            </div>
          )}
        </div>

        {!status || status.recent_blocks.length === 0 ? (
          <div className="text-sm text-slate-600 text-center py-8">
            Waiting for blocks...
          </div>
        ) : (
          <div className="divide-y divide-white/5">
            {status.recent_blocks.map((b) => (
              <div
                key={b.number}
                className="flex items-center gap-4 py-3"
              >
                {/* Icon */}
                {b.is_ours ? (
                  b.missed ? (
                    <div className="w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 bg-red-500/10">
                      <X className="w-3.5 h-3.5 text-red-400" />
                    </div>
                  ) : (
                    <div className="w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 bg-emerald-500/15 ring-1 ring-emerald-500/30">
                      <Shield className="w-3.5 h-3.5 text-emerald-400" />
                    </div>
                  )
                ) : (
                  <div className="w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0 bg-slate-700/40">
                    <Check className="w-3.5 h-3.5 text-slate-500" />
                  </div>
                )}

                {/* Block number */}
                <a
                  href={`https://tronscan.org/#/block/${b.number}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="font-mono text-xs text-slate-500 hover:text-blue-400 transition-colors w-16 flex-shrink-0"
                >
                  #{b.number}
                </a>

                {/* Proposer + address */}
                <div className="flex-1 min-w-0">
                  <span className={`text-sm ${b.is_ours ? 'font-semibold text-white' : 'font-medium text-slate-400'}`}>
                    {b.proposer}
                  </span>
                  {b.address && (
                    <span className="ml-2 font-mono text-xs text-slate-600" title={b.address}>
                      {truncateAddress(b.address)}
                    </span>
                  )}
                </div>

                {/* Tags */}
                <div className="flex items-center gap-2">
                  {b.missed && (
                    <span className="text-xs font-medium text-red-400 bg-red-500/10 border border-red-500/20 px-2 py-0.5 rounded-full">
                      missed
                    </span>
                  )}
                </div>

                {/* Time */}
                <span className="text-xs text-slate-600 font-mono w-16 text-right flex-shrink-0">
                  {formatTime(b.timestamp)}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

    </div>
  )
}

function StatCard({
  label,
  value,
  icon,
  valueClass = 'text-slate-100',
  title,
}: {
  label: string
  value: string | number
  icon: React.ReactNode
  valueClass?: string
  title?: string
}) {
  return (
    <div className="bg-[#111827] border border-white/5 rounded-xl p-5">
      <div className="flex items-center justify-between mb-3">
        <span className="text-xs font-semibold uppercase tracking-widest text-slate-500">{label}</span>
        {icon}
      </div>
      <span className={`text-2xl font-bold ${valueClass}`} title={title}>{value}</span>
    </div>
  )
}
