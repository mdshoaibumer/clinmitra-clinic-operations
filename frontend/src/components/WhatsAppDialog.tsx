import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { X, Send, ExternalLink, Smartphone } from 'lucide-react'
import type { WhatsAppMessageResult } from '@/types/api'

interface WhatsAppDialogProps {
  isOpen: boolean
  onClose: () => void
  messageResult: WhatsAppMessageResult | null
  title: string
}

export default function WhatsAppDialog({ isOpen, onClose, messageResult, title }: WhatsAppDialogProps) {
  const [editedMessage, setEditedMessage] = useState('')
  const [sending, setSending] = useState(false)

  // Sync edited message when messageResult changes
  useEffect(() => {
    if (messageResult) {
      setEditedMessage(messageResult.message)
    }
  }, [messageResult])

  if (!isOpen || !messageResult) return null

  const handleSend = async () => {
    setSending(true)
    try {
      // Use the appropriate URL based on WhatsApp Desktop availability
      const url = messageResult.isDesktopPresent
        ? messageResult.whatsAppUrl
        : messageResult.webUrl

      await window.go.handler.WhatsAppHandler.SendViaWhatsApp(url)
    } catch (err) {
      console.error('Failed to open WhatsApp:', err)
    } finally {
      setSending(false)
      handleClose()
    }
  }

  const handleSendWeb = async () => {
    setSending(true)
    try {
      await window.go.handler.WhatsAppHandler.SendViaWhatsApp(messageResult.webUrl)
    } catch (err) {
      console.error('Failed to open WhatsApp Web:', err)
    } finally {
      setSending(false)
      handleClose()
    }
  }

  const handleClose = () => {
    setEditedMessage('')
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-lg mx-4 max-h-[90vh] flex flex-col">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
          <CardTitle className="flex items-center gap-2 text-lg">
            <svg viewBox="0 0 24 24" className="h-5 w-5 fill-green-500" xmlns="http://www.w3.org/2000/svg">
              <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z" />
            </svg>
            {title}
          </CardTitle>
          <Button variant="ghost" size="icon" onClick={handleClose} className="h-8 w-8">
            <X className="h-4 w-4" />
          </Button>
        </CardHeader>
        <CardContent className="flex flex-col gap-4 overflow-hidden">
          {/* Phone number display */}
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Smartphone className="h-4 w-4" />
            <span>To: +{messageResult.phone}</span>
          </div>

          {/* Message preview/edit */}
          <div className="flex-1 overflow-hidden">
            <label className="text-sm font-medium mb-1 block">Message Preview</label>
            <textarea
              className="w-full h-48 p-3 text-sm border rounded-md resize-none bg-green-50 dark:bg-green-950/20 focus:outline-none focus:ring-2 focus:ring-green-500"
              value={editedMessage}
              onChange={(e) => setEditedMessage(e.target.value)}
              readOnly
            />
          </div>

          {/* Action buttons */}
          <div className="flex gap-2 justify-end pt-2 border-t">
            <Button variant="outline" onClick={handleClose}>
              Skip
            </Button>
            {messageResult.isDesktopPresent && (
              <Button
                onClick={handleSend}
                disabled={sending}
                className="bg-green-600 hover:bg-green-700 text-white"
              >
                <Send className="h-4 w-4 mr-2" />
                {sending ? 'Opening...' : 'Send via WhatsApp'}
              </Button>
            )}
            <Button
              onClick={handleSendWeb}
              disabled={sending}
              variant={messageResult.isDesktopPresent ? 'outline' : 'default'}
              className={!messageResult.isDesktopPresent ? 'bg-green-600 hover:bg-green-700 text-white' : ''}
            >
              <ExternalLink className="h-4 w-4 mr-2" />
              {sending ? 'Opening...' : messageResult.isDesktopPresent ? 'Open in Browser' : 'Send via WhatsApp Web'}
            </Button>
          </div>

          {!messageResult.isDesktopPresent && (
            <p className="text-xs text-muted-foreground text-center">
              WhatsApp Desktop not detected. Message will open in WhatsApp Web in your browser.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
