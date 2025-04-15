import { cn } from "@/lib/utils"
import React from "react"

interface CustomHeatmapCardProps {
  children: React.ReactNode
  className?: string
}

const CustomHeatmapCard = React.forwardRef<
  HTMLDivElement,
  CustomHeatmapCardProps
>(({ children, className }, ref) => (
  <div
    ref={ref}
    className={cn(
      "rounded-xl border bg-card px-4 py-2 backdrop-blur-lg transition-all",
      className
    )}
  >
    {children}
  </div>
))

CustomHeatmapCard.displayName = "CustomHeatmapCard"

export default CustomHeatmapCard 