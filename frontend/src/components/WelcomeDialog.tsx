import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog'
import { Button } from './ui/button'
import { Music, Disc3, Radio, Sparkles } from 'lucide-react'

interface WelcomeDialogProps {
  open: boolean
  onClose: () => void
}

export function WelcomeDialog({ open, onClose }: WelcomeDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <Music className="h-8 w-8 text-primary" />
            <DialogTitle className="text-2xl">Welcome to Rocklist!</DialogTitle>
          </div>
          <DialogDescription className="text-base">
            Create amazing playlists for your Rockbox device using data from your favorite music services.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="flex items-start gap-3">
            <Disc3 className="h-5 w-5 text-primary mt-0.5" />
            <div>
              <h4 className="font-medium">Parse Your Library</h4>
              <p className="text-sm text-muted-foreground">
                Connect your Rockbox device and parse its database to get started.
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <Sparkles className="h-5 w-5 text-primary mt-0.5" />
            <div>
              <h4 className="font-medium">Generate Smart Playlists</h4>
              <p className="text-sm text-muted-foreground">
                Create playlists based on top songs, similar artists, or genre tags using Last.fm, Spotify, or MusicBrainz.
              </p>
            </div>
          </div>

          <div className="flex items-start gap-3">
            <Radio className="h-5 w-5 text-primary mt-0.5" />
            <div>
              <h4 className="font-medium">Enjoy Offline</h4>
              <p className="text-sm text-muted-foreground">
                All matched songs are from your local library - perfect for offline listening!
              </p>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button onClick={onClose}>Get Started</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
