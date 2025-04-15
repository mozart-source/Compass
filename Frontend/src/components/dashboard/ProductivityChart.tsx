"use client"

import { TrendingUp } from "lucide-react"
import { Cell, Pie, PieChart } from "recharts"

import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  ChartConfig,
  ChartContainer,
} from "@/components/ui/chart"

// The value we want to display (productivity percentage)
const PRODUCTIVITY_VALUE = 80;

const chartData = [
  { name: "Completed", value: PRODUCTIVITY_VALUE },
  { name: "Remaining", value: 100 - PRODUCTIVITY_VALUE }
];

// Colors for the two slices using CSS variables
const COLORS = ["hsl(var(--primary))", "hsl(var(--muted))"];

const chartConfig = {
  productivity: {
    label: "Productivity",
    color: "hsl(var(--primary))",
  },
} satisfies ChartConfig

export function ChartRadialStacked() {
  return (
    <Card className="flex flex-col rounded-3xl">
      <CardHeader className="items-center pb-0">
        <CardTitle>Productivity Score</CardTitle>
        <CardDescription>Today's Performance</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-1 items-center pb-0">
        <ChartContainer
          config={chartConfig}
          className="mx-auto aspect-square w-full max-w-[200px]"
        >
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              startAngle={180}
              endAngle={0}
              innerRadius={60}
              outerRadius={80}
              paddingAngle={0}
              dataKey="value"
              cornerRadius={2}
              stroke="none"
            >
              {chartData.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            {/* Central text display */}
            <text x="50%" y="50%" textAnchor="middle">
              <tspan
                x="50%"
                y="45%"
                className="fill-foreground text-2xl font-bold"
                style={{ fontWeight: 'bold', fontSize: '24px' }}
              >
                {PRODUCTIVITY_VALUE}%
              </tspan>
              <tspan
                x="50%"
                y="55%"
                className="fill-muted-foreground text-sm"
                style={{ fontSize: '14px', fill: 'hsl(var(--muted-foreground))' }}
              >
                Productivity
              </tspan>
            </text>
          </PieChart>
        </ChartContainer>
      </CardContent>
      <CardFooter className="flex-col gap-2 text-sm -mt-16">
        <div className="flex items-center gap-2 leading-none font-medium">
          <TrendingUp className="h-4 w-4" />
          5.2% above your average
        </div>
        <div className="text-muted-foreground leading-none">
          Based on your focus time and task completion
        </div>
      </CardFooter>
    </Card>
  )
}
