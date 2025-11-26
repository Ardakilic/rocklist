import { useState, useEffect } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from './components/ui/tabs'
import { FetchTab } from './components/FetchTab'
import { GenerateTab } from './components/GenerateTab'
import { SettingsTab } from './components/SettingsTab'
import { WelcomeDialog } from './components/WelcomeDialog'
import { AboutDialog } from './components/AboutDialog'
import { Toaster } from './components/ui/toaster'
import { Music, Settings, Download, Info } from 'lucide-react'

declare global {
  interface Window {
    go: {
      cmd: {
        App: {
          GetAppInfo: () => Promise<Record<string, string>>
          GetConfig: () => Promise<AppConfig>
          SetRockboxPath: (path: string) => Promise<void>
          SetLastFMCredentials: (apiKey: string, apiSecret: string, enabled: boolean) => Promise<void>
          SetSpotifyCredentials: (clientId: string, clientSecret: string, enabled: boolean) => Promise<void>
          SetMusicBrainzCredentials: (userAgent: string, enabled: boolean) => Promise<void>
          ParseDatabase: (usePrefetched: boolean) => Promise<void>
          GetParseStatus: () => Promise<ParseStatus>
          GetLastParsedAt: () => Promise<string | null>
          GeneratePlaylist: (dataSource: string, playlistType: string, artist: string, tag: string, limit: number) => Promise<Playlist>
          GetSongCount: () => Promise<number>
          GetUniqueArtists: () => Promise<string[]>
          GetUniqueGenres: () => Promise<string[]>
          GetAllPlaylists: () => Promise<Playlist[]>
          DeletePlaylist: (id: number) => Promise<void>
          WipeData: () => Promise<void>
          GetLogs: () => Promise<LogEntry[]>
          ClearLogs: () => void
          GetEnabledSources: () => Promise<string[]>
          SelectDirectory: () => Promise<string>
        }
      }
    }
  }
}

export interface AppConfig {
  rockbox_path: string
  last_parsed_at: string | null
  lastfm: {
    enabled: boolean
    api_key: string
    api_secret: string
  }
  spotify: {
    enabled: boolean
    client_id: string
    client_secret: string
  }
  musicbrainz: {
    enabled: boolean
    user_agent: string
  }
}

export interface ParseStatus {
  in_progress: boolean
  started_at: string | null
  completed_at: string | null
  total_songs: number
  processed_songs: number
  error_count: number
  last_error: string | null
}

export interface Playlist {
  ID: number
  name: string
  description: string
  type: string
  data_source: string
  artist: string
  tag: string
  file_path: string
  song_count: number
  generated_at: string
  exported_at: string | null
}

export interface LogEntry {
  time: string
  level: string
  message: string
}

function App() {
  const [showWelcome, setShowWelcome] = useState(true)
  const [showAbout, setShowAbout] = useState(false)
  const [appInfo, setAppInfo] = useState<Record<string, string>>({})

  useEffect(() => {
    // Check if this is the first visit
    const hasVisited = localStorage.getItem('rocklist_visited')
    if (hasVisited) {
      setShowWelcome(false)
    }

    // Load app info
    if (window.go?.cmd?.App) {
      window.go.cmd.App.GetAppInfo().then(setAppInfo).catch(console.error)
    }
  }, [])

  const handleWelcomeClose = () => {
    localStorage.setItem('rocklist_visited', 'true')
    setShowWelcome(false)
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Music className="h-8 w-8 text-primary" />
            <div>
              <h1 className="text-xl font-bold text-foreground">Rocklist</h1>
              <p className="text-sm text-muted-foreground">Playlist Generator for Rockbox</p>
            </div>
          </div>
          <button
            onClick={() => setShowAbout(true)}
            className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
          >
            <Info className="h-5 w-5" />
            <span className="text-sm">About</span>
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="p-6">
        <Tabs defaultValue="fetch" className="w-full">
          <TabsList className="grid w-full max-w-md grid-cols-3 mb-6">
            <TabsTrigger value="fetch" className="flex items-center gap-2">
              <Download className="h-4 w-4" />
              Fetch
            </TabsTrigger>
            <TabsTrigger value="generate" className="flex items-center gap-2">
              <Music className="h-4 w-4" />
              Generate
            </TabsTrigger>
            <TabsTrigger value="settings" className="flex items-center gap-2">
              <Settings className="h-4 w-4" />
              Settings
            </TabsTrigger>
          </TabsList>

          <TabsContent value="fetch">
            <FetchTab />
          </TabsContent>

          <TabsContent value="generate">
            <GenerateTab />
          </TabsContent>

          <TabsContent value="settings">
            <SettingsTab />
          </TabsContent>
        </Tabs>
      </main>

      {/* Dialogs */}
      <WelcomeDialog open={showWelcome} onClose={handleWelcomeClose} />
      <AboutDialog open={showAbout} onClose={() => setShowAbout(false)} appInfo={appInfo} />
      <Toaster />
    </div>
  )
}

export default App
