"use client"

import { type LucideIcon } from "lucide-react"
import { Link } from 'react-router-dom'

import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"

export function NavMain({
  items,
}: {
  items: {
    title: string
    url: string
    icon?: LucideIcon
    isActive?: boolean
    items?: {
      title: string
      url: string
    }[]
  }[]
}) {
  return (
    <SidebarGroup>
      <SidebarGroupLabel className="text-small text-[#949597]">Navigation</SidebarGroupLabel>
      <SidebarMenu>
        {items.map((item) => (
          <SidebarMenuItem key={item.title}>
            <SidebarMenuButton asChild>
              <Link 
                to={item.url} 
                className="group/nav-item flex items-center gap-2 w-full text-[#e3e4e5] hover:text-foreground data-[state=open]:bg-[#151619] data-[state=open]:text-foreground transition-colors text-small font-medium [&:hover]:bg-[#151619] [&.group-hover]:text-foreground"
              >
                {item.icon && (
                  <item.icon 
                    className="h-4 w-4 text-[#949597] group-hover/nav-item:text-[#ffffff] transition-colors" 
                  />
                )}
                <span>{item.title}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  )
}
