import { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Checkbox } from './ui/checkbox'
import { FolderOpen, Play, Loader2 } from 'lucide-react'
import type { LogEntry, ParseStatus } from '../App'

export function FetchTab() {
  const [rockboxPath, setRockboxPath] = useState('')
  const [usePrefetched, setUsePrefetched] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [_parseStatus, _setParseStatus] = useState<ParseStatus | null>(null)
  const [lastParsedAt, setLastParsedAt] = useState<string | null>(null)
  const [songCount, setSongCount] = useState(0)

  useEffect(() => {
    loadInitialData()
    const interval = setInterval(refreshLogs, 2000)
    return () => clearInterval(interval)
  }, [])

  const loadInitialData = async () => {
    if (!window.go?.cmd?.App) return
    
    try {
      const config = await window.go.cmd.App.GetConfig()
      if (config?.rockbox_path) {
        setRockboxPath(config.rockbox_path)
      }
      
      const lastParsed = await window.go.cmd.App.GetLastParsedAt()
      setLastParsedAt(lastParsed)
      
      const count = await window.go.cmd.App.GetSongCount()
      setSongCount(count)
    } catch (error) {
      console.error('Failed to load initial data:', error)
    }
  }

  const refreshLogs = async () => {
    if (!window.go?.cmd?.App) return
    try {
      const newLogs = await window.go.cmd.App.GetLogs()
      setLogs(newLogs)
    } catch (error) {
      console.error('Failed to refresh logs:', error)
    }
  }

  const handleBrowseFolder = async () => {
    if (!window.go?.cmd?.App) return
    
    try {
      const selectedPath = await window.go.cmd.App.SelectDirectory()
      if (selectedPath) {
        setRockboxPath(selectedPath)
      }
    } catch (error) {
      console.error('Failed to open directory dialog:', error)
    }
  }

  const handleParse = async () => {
    if (!window.go?.cmd?.App) return
    
    setIsLoading(true)
    try {
      await window.go.cmd.App.SetRockboxPath(rockboxPath)
      await window.go.cmd.App.ParseDatabase(usePrefetched)
      
      const count = await window.go.cmd.App.GetSongCount()
      setSongCount(count)
      
      const lastParsed = await window.go.cmd.App.GetLastParsedAt()
      setLastParsedAt(lastParsed)
      
      await refreshLogs()
    } catch (error) {
      console.error('Parse failed:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const formatDate = (dateStr: string | null) => {
    if (!dateStr) return 'Never'
    try {
      return new Date(dateStr).toLocaleString()
    } catch {
      return dateStr
    }
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">Parse Rockbox Database</h2>
        
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="rockbox-path">Rockbox Device Path</Label>
            <div className="flex gap-2">
              <Input
                id="rockbox-path"
                placeholder="/Volumes/IPOD or E:\"
                value={rockboxPath}
                onChange={(e) => setRockboxPath(e.target.value)}
                className="flex-1"
              />
              <Button variant="outline" size="icon" onClick={handleBrowseFolder}>
                <FolderOpen className="h-4 w-4" />
              </Button>
            </div>
            <p className="text-sm text-muted-foreground">
              Path where the .rockbox folder is located
            </p>
          </div>

          <div className="flex items-center space-x-2">
            <Checkbox
              id="use-prefetched"
              checked={usePrefetched}
              onCheckedChange={(checked) => setUsePrefetched(checked as boolean)}
            />
            <Label htmlFor="use-prefetched" className="cursor-pointer">
              Use pre-fetched data (skip parsing)
            </Label>
          </div>

          <div className="flex items-center gap-4">
            <Button onClick={handleParse} disabled={isLoading || !rockboxPath}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Parsing...
                </>
              ) : (
                <>
                  <Play className="mr-2 h-4 w-4" />
                  Parse Now!
                </>
              )}
            </Button>
            
            <div className="text-sm text-muted-foreground">
              <span className="font-medium">{songCount}</span> songs in database
              {lastParsedAt && (
                <span className="ml-2">• Last parsed: {formatDate(lastParsedAt)}</span>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Logs Section */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Logs</h3>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => window.go?.cmd?.App?.ClearLogs()}
          >
            Clear
          </Button>
        </div>
        
        <div className="log-container h-64 overflow-y-auto rounded bg-secondary/50 p-3">
          {logs.length === 0 ? (
            <p className="text-muted-foreground text-sm">No logs yet. Parse the database to see activity.</p>
          ) : (
            logs.map((log, index) => (
              <div key={index} className={`log-entry ${log.level}`}>
                <span className="log-time">
                  {new Date(log.time).toLocaleTimeString()}
                </span>
                <span>{log.message}</span>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}
