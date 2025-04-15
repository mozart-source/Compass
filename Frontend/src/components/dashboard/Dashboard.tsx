import { Button } from "@/components/ui/button"
import { useEffect, useState } from "react"
import { Plus, Brain, ActivityIcon, ArrowRight, Eye } from "lucide-react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import HabitHeatmap from "./HabitHeatmap"
import useHabitHeatmap from "@/hooks/useHabitHeatmap"
import TodoForm from "@/components/todo/Components/TodoForm"
import { TodoFormData, TodoStatus } from '@/components/todo/types-todo'
import { useCreateTodo, useTodoLists } from '@/components/todo/hooks'
import authApi, { User } from '@/api/auth'
import { useQuery } from "@tanstack/react-query"
import { ChartRadialStacked } from "./ProductivityChart"
import PieChart from "./PieChart"
import CalendarWidget from "./CalendarWidget"
import { FocusTime } from "./FocusTime"
import { MiddleBanner } from "./MiddleBanner"
import { useNavigate } from "react-router-dom"

interface TaskMetrics {
  completed: number
  total: number
  upcoming: number
}

interface FocusMetrics {
  todayMinutes: number
  weeklyGoal: number
  weeklyProgress: number
}

interface SystemMetrics {
  keyboardUsage: number
  screenTime: number
  focusScore: number
  productivityScore: number
}

interface Meeting {
  id: string
  time: string
  period: 'AM' | 'PM'
  title: string
  hasVideo?: boolean
  type?: string
}

interface DashboardProps {
  view?: 'tasks' | 'calendar' | 'monitoring';
}

export default function Dashboard({ view }: DashboardProps) {
  const navigate = useNavigate()
  const [greeting, setGreeting] = useState<string>("Good day")
  const [currentTime, setCurrentTime] = useState<string>("")
  const [currentDate, setCurrentDate] = useState<string>("")
  const [showTodoForm, setShowTodoForm] = useState(false)
  const [currentListId, setCurrentListId] = useState<string>('')

  // User authentication query
  const { data: user } = useQuery<User>({
    queryKey: ['user'],
    queryFn: async () => {
      const token = localStorage.getItem('token');
      if (!token) throw new Error('No token found');
      return authApi.getMe();
    },
  });

  // Use the custom hooks
  const { data: todoLists = [] } = useTodoLists(user);
  const createTodoMutation = useCreateTodo();

  // Set default list ID on initial load
  useEffect(() => {
    if (todoLists.length > 0 && !currentListId) {
      const defaultList = todoLists.find(list => list.is_default);
      setCurrentListId(defaultList?.id || todoLists[0].id);
    }
  }, [todoLists, currentListId]);

  const handleTodoFormSubmit = (formData: TodoFormData) => {
    if (!user) return;

    const newTodo = {
      user_id: user.id,
      list_id: currentListId === 'default' ? undefined : currentListId,
      title: formData.title,
      description: formData.description,
      status: TodoStatus.PENDING,
      priority: formData.priority,
      is_recurring: formData.is_recurring,
      due_date: formData.due_date?.toISOString() || null,
      reminder_time: formData.reminder_time?.toISOString() || null,
      tags: formData.tags?.reduce((acc, tag) => ({ ...acc, [tag]: {} }), {}),
      is_completed: false,
      linked_task_id: null,
      linked_calendar_event_id: null,
      recurrence_pattern: {}
    };

    createTodoMutation.mutate(newTodo);
    setShowTodoForm(false);
  };
  
  // Use the habit heatmap hook with proper userId
  const { data: heatmapData, loading: heatmapLoading, error: heatmapError } = useHabitHeatmap(user?.id || '')

  // Update greeting based on time of day
  useEffect(() => {
    const updateDateTime = () => {
      const now = new Date()
      const hours = now.getHours()
      if (hours < 12) {
        setGreeting("Morning")
      } else if (hours < 18) {
        setGreeting("Afternoon")
      } else {
        setGreeting("Evening")
      }

      // Format time (HH:MM am/pm)
      setCurrentTime(now.toLocaleTimeString('en-US', { 
        hour: '2-digit', 
        minute: '2-digit',
        hour12: true 
      }))

      // Format date (Month Day, Year)
      setCurrentDate(now.toLocaleDateString('en-US', { 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
      }))
    }

    // Initial update
    updateDateTime()
    
    // Update every minute
    const intervalId = setInterval(updateDateTime, 60000)
    
    return () => clearInterval(intervalId)
  }, [])

  return (
    <>
      <div className="flex flex-1 flex-col gap-4 p-6">
        {/* Dashboard Label */}
        <p className="text-xs uppercase text-muted-foreground tracking-wider">Dashboard</p>
        
        {/* Header with Greeting and Quick Actions */}
        <div className="flex justify-start">
          {/* Greeting Header */}
          <div>
            <h1 className="text-2xl font-bold tracking-tight leading-none">
              {greeting}, {user?.first_name}
            </h1>
            <p className="text-sm text-muted-foreground mt-2 tracking-wide">{currentDate} · {currentTime}</p>
          </div>
          <div className="col-span-4 flex items-center ml-auto">
                      {/* Quick Actions */}
          <div className="flex gap-2">
            <Button 
              variant="outline"
              size="sm"
              className="gap-2"
              onClick={() => setShowTodoForm(true)}
            >
              <Plus className="h-4 w-4" />
              New Todo
            </Button>
            <Button 
              variant="outline"
              size="sm"
              className="gap-2"
              onClick={() => navigate('/Ai')}
            >
              <Eye className="h-4 w-4" />
              IRIS
            </Button>
            <Dialog>
              <DialogTrigger asChild>
                <Button 
                  variant="outline"
                  size="sm"
                  className="gap-2"
                >
                  <ActivityIcon className="h-4 w-4" />
                  System Status
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[425px]">
                <DialogHeader>
                  <DialogTitle>System Status</DialogTitle>
                </DialogHeader>
                <div className="space-y-4 pt-4">
                  <div className="flex items-center justify-between rounded-lg bg-green-900 p-3">
                    <span className="flex items-center gap-2">
                      <div className="rounded-full p-1.5">
                        <ActivityIcon className="h-4 w-4" />
                      </div>
                      Vision Module
                    </span>
                    <span className="flex items-center gap-1.5 text-green-500">
                      <span className="h-2 w-2 rounded-full bg-green-500"></span>
                      Active
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-lg bg-green-900 p-3">
                    <span className="flex items-center gap-2">
                      <div className="rounded-full p-1.5">
                        <Brain className="h-4 w-4" />
                      </div>
                      Audio Module
                    </span>
                    <span className="flex items-center gap-1.5 text-green-500">
                      <span className="h-2 w-2 rounded-full bg-green-500"></span>
                      Active
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-lg bg-green-900 p-3">
                    <span className="flex items-center gap-2">
                      <div className="rounded-full p-1.5">
                        <Brain className="h-4 w-4" />
                      </div>
                      RAG System
                    </span>
                    <span className="flex items-center gap-1.5 text-green-500">
                      <span className="h-2 w-2 rounded-full bg-green-500"></span>
                      Active
                    </span>
                  </div>
                  <div className="flex items-center justify-between rounded-lg bg-green-900 p-3">
                    <span className="flex items-center gap-2">
                      <div className="rounded-full p-1.5">
                        <Brain className="h-4 w-4" />
                      </div>
                      Agent Ecosystem
                    </span>
                    <span className="flex items-center gap-1.5 text-green-500">
                      <span className="h-2 w-2 rounded-full bg-green-500"></span>
                      Active
                    </span>
                  </div>
                </div>
              </DialogContent>
            </Dialog>
          </div>
          </div>
        </div>

        {/* Habit Heatmap and Today's Meetings */}
        <div className="flex justify-between items-start gap-4">
          {/* Habit Heatmap */}
          <div>
            <HabitHeatmap 
              data={heatmapData}
              loading={heatmapLoading}
              error={heatmapError}
            />
            <Button 
              variant="outline" 
              className="mt-2 bg-[#18191b] rounded-2xl gap-2 w-[230px] h-[52px]"
              onClick={() => navigate('/Todos&Habits')}
            >
              Navigate to Daily Habits <ArrowRight className="h-4 w-4" />
            </Button>
          </div>


          {/* Middle Banner */}
          <div className="flex-1 w-[700px]">
            <MiddleBanner />
          </div>


          {/* Calendar Widget */}
          <div className="">
            <CalendarWidget />
          </div>

        </div>

        {/* Main Grid */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">

          {/* Focus Time/Score */}
          <FocusTime />

          {/* Piechart Overview */}
          <PieChart />

          {/* Productivity Score */}
          <ChartRadialStacked />
        </div>

        {showTodoForm && user && (
          <TodoForm
            onClose={() => setShowTodoForm(false)}
            user={user}
            onSubmit={handleTodoFormSubmit}
            onDelete={() => {}}
            currentListId={currentListId || 'default'}
            listId={currentListId || 'default'}
          />
        )}
      </div>
    </>
  )
}
