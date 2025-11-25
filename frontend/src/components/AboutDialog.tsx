import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from './ui/dialog'
import { Music, Github, Mail } from 'lucide-react'

interface AboutDialogProps {
  open: boolean
  onClose: () => void
  appInfo: Record<string, string>
}

export function AboutDialog({ open, onClose, appInfo }: AboutDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <Music className="h-8 w-8 text-primary" />
            <div>
              <DialogTitle className="text-2xl">{appInfo.name || 'Rocklist'}</DialogTitle>
              <p className="text-sm text-muted-foreground">Version {appInfo.version || '1.0.0'}</p>
            </div>
          </div>
          <DialogDescription>
            {appInfo.description || 'A tool for creating playlists for Rockbox firmware devices'}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div>
            <h4 className="font-medium mb-2">Author</h4>
            <p className="text-sm text-muted-foreground">
              {appInfo.author || 'Arda Kılıçdağı'}
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Mail className="h-4 w-4 text-muted-foreground" />
            <a 
              href={`mailto:${appInfo.email || 'arda@kilicdagi.com'}`}
              className="text-sm text-primary hover:underline"
            >
              {appInfo.email || 'arda@kilicdagi.com'}
            </a>
          </div>

          <div className="flex items-center gap-2">
            <Github className="h-4 w-4 text-muted-foreground" />
            <a 
              href={appInfo.repository || 'https://github.com/Ardakilic/rocklist'}
              target="_blank"
              rel="noopener noreferrer"
              className="text-sm text-primary hover:underline"
            >
              {appInfo.repository || 'github.com/Ardakilic/rocklist'}
            </a>
          </div>

          <div className="pt-4 border-t">
            <p className="text-sm text-muted-foreground">
              Licensed under {appInfo.license || 'MIT'}
            </p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
