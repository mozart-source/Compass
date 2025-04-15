"use client"

import { Bell, ChevronsUpDown, CreditCard, LogOut, Settings, Sparkles } from "lucide-react"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { DropdownMenu, DropdownMenuContent, DropdownMenuGroup, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from "@/components/ui/sidebar"
import { useNavigate } from "react-router-dom"
import { useAuth } from '@/hooks/useAuth'
import { useQueryClient } from '@tanstack/react-query'
import SettingsForm from "@/components/settings-form"
import { useState } from "react"

interface DefaultUser {
  name: string
  email: string
  avatar_url: string
}

const FALLBACK_USER: DefaultUser = {
  name: 'Anonymous',
  email: 'anonymous@example.com',
  avatar_url: ''
}

interface NavUserProps {
  defaultUser?: DefaultUser
  onLogout?: () => void
}

export function NavUser({ defaultUser, onLogout }: NavUserProps) {
  const { isMobile } = useSidebar()
  const { logout, user, isLoadingUser } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [showSettings, setShowSettings] = useState(false)

  // If loading and no defaultUser, show loading state
  if (isLoadingUser && !defaultUser) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton
            size="lg"
            className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
          >
            <div className="flex aspect-square size-8 items-center justify-center">
              <div className="animate-pulse bg-muted rounded-lg h-full w-full" />
            </div>
            <div className="grid flex-1 gap-1">
              <div className="h-4 w-24 animate-pulse bg-muted rounded" />
              <div className="h-3 w-32 animate-pulse bg-muted rounded" />
            </div>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    )
  }

  const handleLogout = () => {
    logout.mutate(undefined, {
      onSuccess: () => {
        queryClient.clear()
        navigate('/login', { replace: true })
      }
    })
  }

  const getInitials = () => {
    if (user) {
      return `${user.first_name[0]}${user.last_name[0]}`.toUpperCase()
    }
    return (defaultUser || FALLBACK_USER).name.slice(0, 2).toUpperCase()
  }

  const getDisplayName = () => {
    if (user) {
      return `${user.first_name} ${user.last_name}`.trim() || user.username
    }
    return (defaultUser || FALLBACK_USER).name
  }

  const getAvatarUrl = () => {
    if (user?.avatar_url) return user.avatar_url
    return (defaultUser || FALLBACK_USER).avatar_url
  }

  const getEmail = () => {
    if (user?.email) return user.email
    return (defaultUser || FALLBACK_USER).email
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
            >
              <div className="flex aspect-square size-8 items-center justify-center">
                <Avatar className="h-full w-full rounded-lg">
                  <AvatarImage src={getAvatarUrl()} alt={getDisplayName()} />
                  <AvatarFallback className="rounded-lg">
                    {getInitials()}
                  </AvatarFallback>
                </Avatar>
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold text-[#e3e4e5] group-hover/nav-item:text-foreground">{getDisplayName()}</span>
                <span className="truncate text-xs text-[#949597]">{getEmail()}</span>
              </div>
              <ChevronsUpDown className="ml-auto size-4 text-[#949597] group-hover/nav-item:text-[#ffffff] transition-colors" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={4}
          >
            <DropdownMenuLabel className="p-0 font-normal">
              <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                <div className="flex aspect-square size-8 items-center justify-center">
                  <Avatar className="h-full w-full rounded-lg">
                    <AvatarImage src={getAvatarUrl()} alt={getDisplayName()} />
                    <AvatarFallback className="rounded-lg">
                      {getInitials()}
                    </AvatarFallback>
                  </Avatar>
                </div>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">{getDisplayName()}</span>
                  <span className="truncate text-xs text-[#949597]">{getEmail()}</span>
                </div>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem>
                <Sparkles className="text-[#949597]" />
                <span className="text-[#e3e4e5]">Upgrade to Pro</span>
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem onSelect={() => setShowSettings(true)}>
                <Settings className="text-[#949597]" />
                <span className="text-[#e3e4e5]">Settings</span>
              </DropdownMenuItem>
              <DropdownMenuItem>
                <CreditCard className="text-[#949597]" />
                <span className="text-[#e3e4e5]">Billing</span>
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Bell className="text-[#949597]" />
                <span className="text-[#e3e4e5]">Notifications</span>
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleLogout}>
              <LogOut className="text-[#949597]"/>
              <span className="text-[#e3e4e5]">Log out</span>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
      {showSettings && <SettingsForm onClose={() => setShowSettings(false)} />}
    </SidebarMenu>
  )
}
