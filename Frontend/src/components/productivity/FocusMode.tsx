import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import { Timer, BarChart2, Brain, Play, Pause } from "lucide-react"
import { useState } from "react"

interface FocusModeProps {
  view?: 'focus' | 'metrics' | 'wellness'
}

export default function FocusMode({ view = 'focus' }: FocusModeProps) {
  const [isTimerActive, setIsTimerActive] = useState(false)
  const [focusMinutes, setFocusMinutes] = useState(25)
  const [progress, setProgress] = useState(0)

  return (
        <div className="flex flex-1 flex-col gap-4 p-6">
          <div className="flex gap-4">
            <Button variant="outline" className="flex items-center gap-2">
              <Timer className="h-4 w-4" />
              Start Session
            </Button>
            <Button variant="outline" className="flex items-center gap-2">
              <BarChart2 className="h-4 w-4" />
              View Stats
            </Button>
            <Button variant="outline" className="flex items-center gap-2">
              <Brain className="h-4 w-4" />
              AI Insights
            </Button>
          </div>

          <Tabs defaultValue={view} className="space-y-4">
            <TabsList>
              <TabsTrigger value="focus">Focus Mode</TabsTrigger>
              <TabsTrigger value="metrics">Metrics</TabsTrigger>
              <TabsTrigger value="wellness">Wellness</TabsTrigger>
            </TabsList>

            <TabsContent value="focus" className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                <Card>
                  <CardHeader>
                    <CardTitle>Pomodoro Timer</CardTitle>
                    <CardDescription>Stay focused with timed work sessions</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="flex justify-center">
                      <div className="w-48 h-48 rounded-full border-4 border-primary flex items-center justify-center">
                        <div className="text-4xl font-bold">{focusMinutes}:00</div>
                      </div>
                    </div>
                    <div className="flex justify-center space-x-2">
                      <Button
                        variant={isTimerActive ? "destructive" : "default"}
                        onClick={() => setIsTimerActive(!isTimerActive)}
                      >
                        {isTimerActive ? (
                          <><Pause className="mr-2 h-4 w-4" /> Pause</>
                        ) : (
                          <><Play className="mr-2 h-4 w-4" /> Start</>
                        )}
                      </Button>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>Daily Progress</CardTitle>
                    <CardDescription>Your focus time today</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="space-y-2">
                      <div className="flex justify-between">
                        <span>Daily Focus Time</span>
                        <span className="font-medium">2h 15m</span>
                      </div>
                      <Progress value={45} />
                      <p className="text-xs text-muted-foreground">
                        45% of daily goal (5 hours)
                      </p>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            <TabsContent value="metrics" className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                <Card>
                  <CardHeader>
                    <CardTitle>Focus Analytics</CardTitle>
                    <CardDescription>Track your productivity metrics</CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="space-y-2">
                      <div className="flex justify-between">
                        <span>Daily Focus Time</span>
                        <span className="font-medium">2h 15m</span>
                      </div>
                      <Progress value={45} />
                      <p className="text-xs text-muted-foreground">
                        45% of daily goal (5 hours)
                      </p>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            <TabsContent value="wellness" className="space-y-4">
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Stress Level
                    </CardTitle>
                    <Brain className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">Low</div>
                    <Progress value={30} className="mt-2" />
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Break Time
                    </CardTitle>
                    <Timer className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">45m</div>
                    <p className="text-xs text-muted-foreground">
                      Today's total breaks
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Productivity Score
                    </CardTitle>
                    <BarChart2 className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">85%</div>
                    <Progress value={85} className="mt-2" />
                  </CardContent>
                </Card>
              </div>
            </TabsContent>
          </Tabs>
        </div>
  )
}
