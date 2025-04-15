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

export function NavProjects({
  projects,
}: {
  projects: {
    name: string
    url: string
    icon: LucideIcon
  }[]
}) {
  return (
    <SidebarGroup className="group-data-[collapsible=icon]:hidden">
      <SidebarGroupLabel className="text-small text-[#949597]">Projects</SidebarGroupLabel>
      <SidebarMenu>
        {projects.map((item) => (
          <SidebarMenuItem key={item.name}>
            <SidebarMenuButton asChild>
              <Link 
                to={item.url} 
                className="group/nav-item flex items-center gap-2 w-full text-[#e3e4e5] hover:text-foreground data-[state=open]:bg-[#151619] data-[state=open]:text-foreground transition-colors text-small font-medium [&:hover]:bg-[#151619] [&.group-hover]:text-foreground"
              >
                <item.icon 
                  className="h-4 w-4 text-[#949597] group-hover/nav-item:text-[#ffffff] transition-colors" 
                />
                <span>{item.name}</span>
              </Link>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  )
}
