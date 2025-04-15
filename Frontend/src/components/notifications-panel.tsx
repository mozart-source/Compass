import * as React from "react"
import { Bell, X, Check, Calendar, Users, FileText, User, RefreshCw, PlusCircle, Settings, HelpCircle } from "lucide-react"
import { useWebSocket } from "@/hooks/useWebSocket"
import { cn } from "@/lib/utils"
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "@/components/ui/sidebar"

export function NotificationsPanel() {
  const [isOpen, setIsOpen] = React.useState(false)
  const [activeTab, setActiveTab] = React.useState("all")
  const panelRef = React.useRef<HTMLDivElement>(null)
  const { notifications, markAsRead, clearNotifications } = useWebSocket(
    localStorage.getItem("token")
  )

  const unreadCount = notifications.filter((n) => !n.read).length
  const displayNotifications = activeTab === "all" 
    ? notifications 
    : notifications.filter(n => !n.read)

  // Group notifications by date
  const today = new Date().toDateString()
  const todayNotifications = displayNotifications.filter(
    n => new Date(n.createdAt).toDateString() === today
  )
  const tomorrowNotifications = displayNotifications.filter(
    n => {
      const date = new Date(n.createdAt)
      const tomorrow = new Date()
      tomorrow.setDate(tomorrow.getDate() + 1)
      return date.toDateString() === tomorrow.toDateString()
    }
  )
  const otherNotifications = displayNotifications.filter(
    n => {
      const date = new Date(n.createdAt)
      const tomorrow = new Date()
      tomorrow.setDate(tomorrow.getDate() + 1)
      return date.toDateString() !== today && date.toDateString() !== tomorrow.toDateString()
    }
  )

  // Close panel when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (panelRef.current && !panelRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener("mousedown", handleClickOutside)
    return () => document.removeEventListener("mousedown", handleClickOutside)
  }, [])

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <div className="relative" ref={panelRef}>
          <SidebarMenuButton
            onClick={() => setIsOpen(!isOpen)}
            className={cn(
              "group/nav-item flex items-center gap-2 w-full text-[#e3e4e5] hover:text-foreground hover:bg-[#151619] transition-colors text-small font-medium"
            )}
            aria-label="Notifications"
          >
            <Bell className="h-4 w-4 text-[#949597] group-hover/nav-item:text-[#ffffff] transition-colors" />
            <span>Notifications</span>
            {unreadCount > 0 && (
              <div className="ml-auto">
                <span className="flex items-center justify-center h-5 w-5 text-[11px] font-medium text-red-100 bg-red-600 rounded-full">
                  {unreadCount}
                </span>
              </div>
            )}
          </SidebarMenuButton>

          {isOpen && (
            <div 
              className="absolute left-full ml-2 bottom-0 w-96 bg-popover rounded-lg shadow-lg overflow-hidden z-50 border border-border animate-in fade-in-0 zoom-in-95"
              style={{ transformOrigin: "center center" }}
            >
              <div className="p-4 border-b flex justify-between items-center">
                <h3 className="text-base font-semibold">Notifications</h3>
                <div className="flex space-x-1">
                  <button 
                    onClick={() => setIsOpen(false)}
                    className="p-1 rounded-full hover:bg-muted transition-colors"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
              </div>
              
              <div className="flex border-b">
                <button 
                  onClick={() => setActiveTab("all")}
                  className={cn(
                    "flex-1 py-2 text-sm font-medium text-center transition-colors",
                    activeTab === "all" 
                      ? "text-foreground border-b-2 border-foreground" 
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  All
                </button>
                <button 
                  onClick={() => setActiveTab("unread")}
                  className={cn(
                    "flex-1 py-2 text-sm font-medium text-center transition-colors",
                    activeTab === "unread" 
                      ? "text-foreground border-b-2 border-foreground" 
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  Unread ({unreadCount})
                </button>
              </div>

              <div className="max-h-[60vh] overflow-y-auto">
                {displayNotifications.length === 0 ? (
                  <div className="p-4 text-center text-muted-foreground">
                    No notifications
                  </div>
                ) : (
                  <>
                    {todayNotifications.length > 0 && (
                      <div>
                        <div className="px-4 py-2 bg-muted/50 text-xs font-medium text-muted-foreground">
                          Today
                        </div>
                        {todayNotifications.map((notification) => (
                          <NotificationItem 
                            key={notification.id} 
                            notification={notification} 
                            markAsRead={markAsRead} 
                          />
                        ))}
                      </div>
                    )}
                    
                    {tomorrowNotifications.length > 0 && (
                      <div>
                        <div className="px-4 py-2 bg-muted/50 text-xs font-medium text-muted-foreground">
                          Tomorrow
                        </div>
                        {tomorrowNotifications.map((notification) => (
                          <NotificationItem 
                            key={notification.id} 
                            notification={notification} 
                            markAsRead={markAsRead} 
                          />
                        ))}
                      </div>
                    )}
                    
                    {otherNotifications.length > 0 && (
                      <div>
                        <div className="px-4 py-2 bg-muted/50 text-xs font-medium text-muted-foreground">
                          Other
                        </div>
                        {otherNotifications.map((notification) => (
                          <NotificationItem 
                            key={notification.id} 
                            notification={notification} 
                            markAsRead={markAsRead} 
                          />
                        ))}
                      </div>
                    )}
                  </>
                )}
              </div>
              
              {displayNotifications.length > 0 && (
                <div className="p-2 border-t text-center">
                  <button
                    onClick={clearNotifications}
                    className="text-sm text-foreground hover:text-foreground/90 transition-colors"
                  >
                    Mark all as read
                  </button>
                </div>
              )}
            </div>
          )}
        </div>
      </SidebarMenuItem>
      <SidebarMenuItem>
        <SidebarMenuButton
          className="group/nav-item flex items-center gap-2 w-full text-[#e3e4e5] hover:text-foreground hover:bg-[#151619] transition-colors text-small font-medium"
          aria-label="Settings"
        >
          <Settings className="h-4 w-4 text-[#949597] group-hover/nav-item:text-[#ffffff] transition-colors" />
          <span>Settings</span>
        </SidebarMenuButton>
      </SidebarMenuItem>
      <SidebarMenuItem>
        <SidebarMenuButton
          className="group/nav-item flex items-center gap-2 w-full text-[#e3e4e5] hover:text-foreground hover:bg-[#151619] transition-colors text-small font-medium"
          aria-label="Get Help"
        >
          <HelpCircle className="h-4 w-4 text-[#949597] group-hover/nav-item:text-[#ffffff] transition-colors" />
          <span>Get Help</span>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}

interface NotificationItemProps {
  notification: {
    id: string
    title: string
    message?: string
    content?: string
    read: boolean
    createdAt: string
    type?: string
    status?: string
  }
  markAsRead: (id: string) => void
}

function NotificationItem({ notification, markAsRead }: NotificationItemProps) {
  // Function to determine which icon to show based on notification content
  const getNotificationIcon = () => {
    const content = notification.title?.toLowerCase() || '';
    if (content.includes('meeting')) return Calendar;
    if (content.includes('contract')) return FileText;
    if (content.includes('team')) return Users;
    if (content.includes('added')) return PlusCircle;
    if (content.includes('renewal')) return RefreshCw;
    return User; // Default icon
  };
  
  const NotificationIcon = getNotificationIcon();
  
  // Get message content from either message or content field
  const messageContent = notification.message || notification.content || '';
  
  return (
    <div
      className={cn(
        "p-4 border-b hover:bg-muted/50 transition-colors",
        !notification.read && "bg-primary/5"
      )}
    >
      <div className="flex items-start">
        <div className="flex-shrink-0 mr-3">
          <div className="h-9 w-9 rounded-full bg-primary/10 flex items-center justify-center">
            <NotificationIcon className="h-4 w-4 text-foreground" />
          </div>
        </div>
        <div className="flex-1">
          <p className="font-medium text-sm">{notification.title}</p>
          <p className="text-sm text-muted-foreground mt-1">
            {messageContent}
          </p>
          <p className="text-xs text-muted-foreground/70 mt-1">
            {new Date(notification.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </p>
        </div>
        {!notification.read && (
          <button
            onClick={() => markAsRead(notification.id)}
            className="ml-2 text-primary hover:text-primary/90 transition-colors"
          >
            <Check className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  )
} 