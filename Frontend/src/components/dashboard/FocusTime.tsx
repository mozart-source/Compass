"use client";
import { Card } from "@/components/ui/card";
import { Focus, ClockFading, Settings } from "lucide-react";
import { Progress } from "@/components/ui/progress";
import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "@/components/ui/use-toast";
import { useFocusSettings } from "@/hooks/useFocusSettings";

interface DailyFocus {
  day: string;
  minutes: number;
}

export function FocusTime() {
  const {
    data: focusData,
    loading,
    isConnected,
    refreshFocus,
    updateSettings,
  } = useFocusSettings();

  const [focusScore, setFocusScore] = useState(0);
  const [focusTime, setFocusTime] = useState({ hours: 0, minutes: 0 });
  const [dailyBreakdown, setDailyBreakdown] = useState<DailyFocus[]>([]);
  const [streak, setStreak] = useState(0);
  const [targetHours, setTargetHours] = useState(4); // Default to 4 hours
  const [newTargetHours, setNewTargetHours] = useState("4");
  const [isUpdating, setIsUpdating] = useState(false);

  useEffect(() => {
    if (focusData) {
      // Get user's configured target or default to 4 hours (14400 seconds)
      const targetDailySeconds = Number(
        focusData.daily_target_seconds || 4 * 60 * 60
      );
      setTargetHours(targetDailySeconds / 3600);
      setNewTargetHours(String(targetDailySeconds / 3600));

      // Calculate focus score (0-100)
      const totalSeconds = Number(focusData.total_focus_seconds) || 0;
      const calculatedScore = Math.min(
        Math.round((totalSeconds / targetDailySeconds) * 100),
        100
      );
      setFocusScore(calculatedScore || 0);

      // Format focus time
      const hours = Math.floor(totalSeconds / 3600);
      const minutes = Math.floor((totalSeconds % 3600) / 60);
      setFocusTime({ hours, minutes });

      // Set streak
      setStreak(Number(focusData.streak) || 0);

      // Set daily breakdown data
      if (
        focusData.daily_breakdown &&
        Array.isArray(focusData.daily_breakdown)
      ) {
        setDailyBreakdown(focusData.daily_breakdown);
      }
    }
  }, [focusData]);

  // Format focus time for display
  const formattedFocusTime = `${focusTime.hours}H ${focusTime.minutes}M`;

  // Format streak message
  const streakMessage =
    streak > 0
      ? `${streak} day${streak > 1 ? "s" : ""} streak!`
      : "Start your streak today!";

  // Calculate progress percentage for the progress bar based on user's target
  const targetMinutes = targetHours * 60;
  const focusTimePercentage = Math.min(
    ((focusTime.hours * 60 + focusTime.minutes) / targetMinutes) * 100,
    100
  );

  const handleUpdateFocusTarget = async () => {
    try {
      setIsUpdating(true);

      const hours = parseFloat(newTargetHours);
      if (isNaN(hours) || hours <= 0 || hours > 24) {
        toast({
          title: "Invalid target",
          description: "Please enter a valid number of hours between 0 and 24",
          variant: "destructive",
        });
        return;
      }

      // Convert hours to seconds
      const dailyTargetSeconds = Math.round(hours * 3600);

      await updateSettings({ daily_target_seconds: dailyTargetSeconds });

      toast({
        title: "Settings updated",
        description: `Your daily focus target is now ${hours} hours`,
      });
    } catch (error) {
      console.error("Error updating focus target:", error);
      toast({
        title: "Error",
        description: "Failed to update focus target",
        variant: "destructive",
      });
    } finally {
      setIsUpdating(false);
    }
  };

  return (
    <Card className="flex flex-col w-auto rounded-3xl p-5 shadow-lg">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-xl font-medium">Today's Progress</h3>
          <p className="text-sm text-muted-foreground">{streakMessage}</p>
        </div>
        <div className="flex items-center gap-2">
          {/* Connection indicator */}
          <div
            className={`w-2 h-2 rounded-full ${
              isConnected ? "bg-green-500" : "bg-gray-400"
            }`}
            title={isConnected ? "Connected" : "Disconnected"}
          />

          {/* Refresh button */}
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={refreshFocus}
            title="Refresh focus data"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"></path>
              <path d="M21 3v5h-5"></path>
              <path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"></path>
              <path d="M8 16H3v5"></path>
            </svg>
          </Button>

          {/* Settings dialog */}
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <Settings className="h-4 w-4" />
              </Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Focus Settings</DialogTitle>
              </DialogHeader>
              <div className="py-4">
                <Label htmlFor="targetHours">Daily Focus Target (hours)</Label>
                <Input
                  id="targetHours"
                  type="number"
                  min="0.5"
                  max="24"
                  step="0.5"
                  value={newTargetHours}
                  onChange={(e) => setNewTargetHours(e.target.value)}
                />
              </div>
              <DialogFooter>
                <DialogClose asChild>
                  <Button variant="outline">Cancel</Button>
                </DialogClose>
                <DialogClose asChild>
                  <Button
                    onClick={handleUpdateFocusTarget}
                    disabled={isUpdating}
                  >
                    {isUpdating ? "Updating..." : "Save Changes"}
                  </Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <Focus className="h-4 w-4 text-blue-500" />
            <span className="text-sm text-muted-foreground">Focus Score</span>
          </div>
          <div className="text-2xl font-bold">{focusScore}%</div>
          <Progress value={focusScore} className="h-1.5" />
        </div>
        <div className="space-y-2">
          <div className="flex items-center gap-2">
            <ClockFading className="h-4 w-4 text-blue-500" />
            <span className="text-sm text-muted-foreground">Focus Time</span>
          </div>
          <div className="text-2xl font-bold">{formattedFocusTime}</div>
          <Progress value={focusTimePercentage} className="h-1.5" />
        </div>
      </div>

      <div className="relative h-[100px] mt-auto">
        {loading ? (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="animate-pulse text-sm text-zinc-500">
              Loading focus data...
            </div>
          </div>
        ) : (
          <div className="absolute inset-0 flex items-end">
            {/* Display daily focus activity from the backend data */}
            {dailyBreakdown.length > 0
              ? dailyBreakdown.map((day, i) => {
                  // Convert minutes to height percentage (max height for 120 minutes)
                  const heightPercentage = Math.min(
                    (day.minutes / 120) * 100,
                    100
                  );

                  return (
                    <div
                      key={i}
                      className="flex-1 mx-0.5"
                      style={{ height: `${heightPercentage}%` }}
                    >
                      <div className="w-full h-full rounded-t-sm bg-gradient-to-t from-blue-500/50 to-blue-500/20" />
                    </div>
                  );
                })
              : // Fallback to static data if no daily breakdown is available
                [40, 65, 45, 80, 55, 85, 60].map((height, i) => (
                  <div
                    key={i}
                    className="flex-1 mx-0.5"
                    style={{ height: `${height}%` }}
                  >
                    <div className="w-full h-full rounded-t-sm bg-gradient-to-t from-blue-500/50 to-blue-500/20" />
                  </div>
                ))}
          </div>
        )}
      </div>

      {/* Target indicator */}
      <div className="mt-4 text-xs text-muted-foreground text-right">
        <span>Target: {targetHours} hours/day</span>
      </div>
    </Card>
  );
}
