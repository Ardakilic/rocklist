import { useState, useEffect } from 'react'
import { Button } from './ui/button'
import { Input } from './ui/input'
import { Label } from './ui/label'
import { Checkbox } from './ui/checkbox'
import { Save, Trash2, AlertTriangle } from 'lucide-react'
import type { AppConfig } from '../App'

export function SettingsTab() {
  const [config, setConfig] = useState<AppConfig | null>(null)
  const [showWipeConfirm, setShowWipeConfirm] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    if (!window.go?.cmd?.App) return
    try {
      const cfg = await window.go.cmd.App.GetConfig()
      setConfig(cfg)
    } catch (error) {
      console.error('Failed to load config:', error)
    }
  }

  const handleSave = async () => {
    if (!window.go?.cmd?.App || !config) return
    
    setIsSaving(true)
    try {
      await window.go.cmd.App.SetLastFMCredentials(
        config.lastfm.api_key,
        config.lastfm.api_secret,
        config.lastfm.enabled
      )
      await window.go.cmd.App.SetSpotifyCredentials(
        config.spotify.client_id,
        config.spotify.client_secret,
        config.spotify.enabled
      )
      await window.go.cmd.App.SetMusicBrainzCredentials(
        config.musicbrainz.user_agent,
        config.musicbrainz.enabled
      )
    } catch (error) {
      console.error('Failed to save config:', error)
    } finally {
      setIsSaving(false)
    }
  }

  const handleWipeData = async () => {
    if (!window.go?.cmd?.App) return
    try {
      await window.go.cmd.App.WipeData()
      setShowWipeConfirm(false)
    } catch (error) {
      console.error('Failed to wipe data:', error)
    }
  }

  if (!config) {
    return <div className="text-muted-foreground">Loading settings...</div>
  }

  return (
    <div className="space-y-6">
      {/* Last.fm Settings */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Last.fm</h3>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="lastfm-enabled"
              checked={config.lastfm.enabled}
              onCheckedChange={(checked) => 
                setConfig({ ...config, lastfm: { ...config.lastfm, enabled: checked as boolean } })
              }
            />
            <Label htmlFor="lastfm-enabled">Enabled</Label>
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>API Key</Label>
            <Input
              type="password"
              value={config.lastfm.api_key}
              onChange={(e) => setConfig({ ...config, lastfm: { ...config.lastfm, api_key: e.target.value } })}
              placeholder="Your Last.fm API key"
            />
          </div>
          <div className="space-y-2">
            <Label>API Secret</Label>
            <Input
              type="password"
              value={config.lastfm.api_secret}
              onChange={(e) => setConfig({ ...config, lastfm: { ...config.lastfm, api_secret: e.target.value } })}
              placeholder="Your Last.fm API secret"
            />
          </div>
        </div>
        <p className="text-sm text-muted-foreground mt-2">
          Get your API key at <a href="https://www.last.fm/api/account/create" target="_blank" className="text-primary hover:underline">last.fm/api</a>
        </p>
      </div>

      {/* Spotify Settings */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">Spotify</h3>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="spotify-enabled"
              checked={config.spotify.enabled}
              onCheckedChange={(checked) => 
                setConfig({ ...config, spotify: { ...config.spotify, enabled: checked as boolean } })
              }
            />
            <Label htmlFor="spotify-enabled">Enabled</Label>
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label>Client ID</Label>
            <Input
              type="password"
              value={config.spotify.client_id}
              onChange={(e) => setConfig({ ...config, spotify: { ...config.spotify, client_id: e.target.value } })}
              placeholder="Your Spotify Client ID"
            />
          </div>
          <div className="space-y-2">
            <Label>Client Secret</Label>
            <Input
              type="password"
              value={config.spotify.client_secret}
              onChange={(e) => setConfig({ ...config, spotify: { ...config.spotify, client_secret: e.target.value } })}
              placeholder="Your Spotify Client Secret"
            />
          </div>
        </div>
        <p className="text-sm text-muted-foreground mt-2">
          Get credentials at <a href="https://developer.spotify.com/dashboard" target="_blank" className="text-primary hover:underline">developer.spotify.com</a>
        </p>
      </div>

      {/* MusicBrainz Settings */}
      <div className="rounded-lg border bg-card p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold">MusicBrainz</h3>
          <div className="flex items-center space-x-2">
            <Checkbox
              id="musicbrainz-enabled"
              checked={config.musicbrainz.enabled}
              onCheckedChange={(checked) => 
                setConfig({ ...config, musicbrainz: { ...config.musicbrainz, enabled: checked as boolean } })
              }
            />
            <Label htmlFor="musicbrainz-enabled">Enabled</Label>
          </div>
        </div>
        <div className="space-y-2">
          <Label>User Agent</Label>
          <Input
            value={config.musicbrainz.user_agent || 'Rocklist/1.0.0 ( https://github.com/Ardakilic/rocklist )'}
            onChange={(e) => setConfig({ ...config, musicbrainz: { ...config.musicbrainz, user_agent: e.target.value } })}
            placeholder="Application/version (contact-url)"
          />
          <p className="text-sm text-muted-foreground">
            MusicBrainz requires a descriptive user agent. Format: AppName/Version (contact-url)
          </p>
        </div>
      </div>

      {/* Save Button */}
      <div className="flex gap-4">
        <Button onClick={handleSave} disabled={isSaving}>
          <Save className="mr-2 h-4 w-4" />
          {isSaving ? 'Saving...' : 'Save Settings'}
        </Button>
      </div>

      {/* Danger Zone */}
      <div className="rounded-lg border border-destructive/50 bg-card p-6">
        <h3 className="text-lg font-semibold text-destructive mb-4">Danger Zone</h3>
        
        {!showWipeConfirm ? (
          <Button variant="destructive" onClick={() => setShowWipeConfirm(true)}>
            <Trash2 className="mr-2 h-4 w-4" />
            Wipe Pre-fetched Data
          </Button>
        ) : (
          <div className="space-y-4">
            <div className="flex items-start gap-2 text-destructive">
              <AlertTriangle className="h-5 w-5 mt-0.5" />
              <div>
                <p className="font-medium">Are you sure?</p>
                <p className="text-sm text-muted-foreground">
                  This will delete all parsed song data and playlists. This cannot be undone.
                </p>
              </div>
            </div>
            <div className="flex gap-2">
              <Button variant="destructive" onClick={handleWipeData}>
                Yes, wipe all data
              </Button>
              <Button variant="outline" onClick={() => setShowWipeConfirm(false)}>
                Cancel
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
