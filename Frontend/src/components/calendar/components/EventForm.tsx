import React, { useState, useEffect } from 'react';
import DatePicker from 'react-datepicker';
import { 
  X, 
  MapPin, 
  Calendar, 
  Clock, 
  Repeat, 
  Tag, 
  AlertCircle, 
  Edit3, 
  Flag,
  CheckSquare,
  Bookmark,
  ChevronDown
} from 'lucide-react';
import { useCreateEvent, useUpdateEvent, useDeleteEvent, useUpdateOccurrenceById } from '@/components/calendar/hooks';
import { 
  CalendarEvent, 
  EventType,
  CreateCalendarEventRequest, 
  UpdateCalendarEventRequest,
  RecurrenceType,
  CreateRecurrenceRuleRequest
} from '../types';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import Checkbox from "@/components/ui/checkbox";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import "react-datepicker/dist/react-datepicker.css";
import './EventForm.css';

interface EventFormProps {
  task?: CalendarEvent | null;
  onClose: () => void;
  userId?: string;
}

// Predefined color options
const colorOptions = [
  { name: 'Blue', value: '#3b82f6' },
  { name: 'Red', value: '#ef4444' },
  { name: 'Green', value: '#10b981' },
  { name: 'Yellow', value: '#f59e0b' },
  { name: 'Purple', value: '#8b5cf6' },
  { name: 'Pink', value: '#ec4899' },
  { name: 'Indigo', value: '#6366f1' },
  { name: 'Gray', value: '#6b7280' },
  { name: 'Teal', value: '#14b8a6' },
  { name: 'Orange', value: '#f97316' },
];

const EventForm: React.FC<EventFormProps> = ({ task, onClose, userId }) => {
  const createEvent = useCreateEvent();
  const updateEvent = useUpdateEvent();
  const updateOccurrenceById = useUpdateOccurrenceById();
  const deleteEvent = useDeleteEvent();
  const [isClosing, setIsClosing] = useState(false);
  const [updateOption, setUpdateOption] = useState<'single' | 'all'>('single');
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showCustomColor, setShowCustomColor] = useState(false);
  
  const [formData, setFormData] = useState<CreateCalendarEventRequest>({
    title: task?.title || '',
    description: task?.description || '',
    event_type: task?.event_type || 'None',
    start_time: task?.start_time ? new Date(task.start_time) : new Date(),
    end_time: task?.end_time ? new Date(task.end_time) : new Date(Date.now() + 60 * 60000),
    is_all_day: task?.is_all_day || false,
    location: task?.location,
    color: task?.color || '#3b82f6',
    transparency: task?.transparency || 'opaque',
  });

  const [recurrenceData, setRecurrenceData] = useState({
    enabled: false,
    freq: 'None' as RecurrenceType,
    interval: 1,
    byDay: [] as string[],
    byMonth: [] as number[],
    byMonthDay: [] as number[],
    until: null as Date | null,
    count: null as number | null,
  });
  
  const isRecurring = !!task?.recurrence_rules?.length || recurrenceData.enabled;
  const isOccurrence = !!task?.is_occurrence;

  const eventTypes: { value: EventType; label: string; icon: React.ReactNode }[] = [
    { value: 'None', label: 'None', icon: <Tag className="h-4 w-4" /> },
    { value: 'Task', label: 'Task', icon: <CheckSquare className="h-4 w-4" /> },
    { value: 'Meeting', label: 'Meeting', icon: <Calendar className="h-4 w-4" /> },
    { value: 'Todo', label: 'Todo', icon: <CheckSquare className="h-4 w-4" /> },
    { value: 'Holiday', label: 'Holiday', icon: <Flag className="h-4 w-4" /> },
    { value: 'Reminder', label: 'Reminder', icon: <AlertCircle className="h-4 w-4" /> }
  ];

  const statusTypes = [
    { value: 'Upcoming', label: 'Upcoming' },
    { value: 'In Progress', label: 'In Progress' },
    { value: 'Completed', label: 'Completed' },
    { value: 'Cancelled', label: 'Cancelled' },
    { value: 'Blocked', label: 'Blocked' },
    { value: 'Under Review', label: 'Under Review' },
    { value: 'Deferred', label: 'Deferred' }
  ];

  const recurrenceTypes: { value: RecurrenceType; label: string }[] = [
    { value: 'None', label: 'None' },
    { value: 'Daily', label: 'Daily' },
    { value: 'Weekly', label: 'Weekly' },
    { value: 'Biweekly', label: 'Biweekly' },
    { value: 'Monthly', label: 'Monthly' },
    { value: 'Yearly', label: 'Yearly' },
    { value: 'Custom', label: 'Custom' }
  ];

  useEffect(() => {
    if (task) {
      setFormData({
        title: task.title,
        description: task.description,
        event_type: task.event_type,
        start_time: new Date(task.start_time),
        end_time: new Date(task.end_time),
        is_all_day: task.is_all_day,
        location: task.location,
        color: task.color || '#3b82f6',
        transparency: task.transparency,
      });
      
      if (!colorOptions.some(c => c.value === task.color)) {
        setShowCustomColor(true);
      }

      if (task.recurrence_rules?.[0]) {
        const rule = task.recurrence_rules[0];
        setRecurrenceData({
          enabled: true,
          freq: rule.freq,
          interval: rule.interval,
          byDay: rule.by_day || [],
          byMonth: rule.by_month?.map(Number) || [],
          byMonthDay: rule.by_month_day?.map(Number) || [],
          until: rule.until ? new Date(rule.until) : null,
          count: rule.count || null,
        });
      }
    }
  }, [task]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // If creating a new event or event is not recurring/occurrence, save directly
    if (!task?.id || (!isRecurring && !isOccurrence)) {
      saveEvent();
      return;
    }
    
    // Otherwise show the update modal for confirmation
    setShowUpdateModal(true);
  };

  const saveEvent = async (updateType?: 'single' | 'all') => {
    setIsClosing(true);
    
    try {
      const eventData: CreateCalendarEventRequest = {
        ...formData,
        recurrence_rule: recurrenceData.enabled ? {
          freq: recurrenceData.freq,
          interval: recurrenceData.interval,
          by_day: recurrenceData.byDay.length > 0 ? recurrenceData.byDay : undefined,
          by_month: recurrenceData.byMonth.length > 0 ? recurrenceData.byMonth : undefined,
          by_month_day: recurrenceData.byMonthDay.length > 0 ? recurrenceData.byMonthDay : undefined,
          until: recurrenceData.until || undefined,
          count: recurrenceData.count || undefined,
        } : undefined
      };

      if (task?.id) {
        const updateData: UpdateCalendarEventRequest = {
          title: eventData.title,
          description: eventData.description,
          event_type: eventData.event_type,
          is_all_day: eventData.is_all_day,
          location: eventData.location,
          color: eventData.color,
          transparency: eventData.transparency,
        };

        if (task.is_occurrence && task.original_event_id && task.occurrence_id) {
          const selectedOption = updateType || updateOption;
          
          if (selectedOption === 'single') {
            await updateOccurrenceById.mutateAsync({
              occurrenceId: task.occurrence_id,
              updates: {
                ...updateData,
                start_time: eventData.start_time,
                end_time: eventData.end_time,
              }
            });
          } else if (selectedOption === 'all') {
            await updateEvent.mutateAsync({
              eventId: task.original_event_id,
              event: {
                ...updateData,
                preserve_date_sequence: true,
                start_time: eventData.start_time,
                end_time: eventData.end_time,
              }
            });
          }
        } else {
          updateData.start_time = eventData.start_time;
          updateData.end_time = eventData.end_time;
          
          await updateEvent.mutateAsync({
            eventId: task.id,
            event: updateData
          });
        }
      } else {
        await createEvent.mutateAsync(eventData);
      }
      setTimeout(onClose, 300);
    } catch (error) {
      console.error('Failed to save event:', error);
      setIsClosing(false);
    }
  };

  const handleDelete = () => {
    if (!task?.id) return;
    
    if (isRecurring || isOccurrence) {
      setShowDeleteModal(true);
    } else {
      deleteEventConfirmed();
    }
  };

  const deleteEventConfirmed = async (deleteType?: 'single' | 'all') => {
    setIsClosing(true);
    
    try {
      if (!task?.id) return;
      
      if (task.is_occurrence && task.original_event_id && deleteType === 'all') {
        await deleteEvent.mutateAsync(task.original_event_id);
      } else {
        await deleteEvent.mutateAsync(task.id);
      }
      
      setTimeout(onClose, 300);
    } catch (error) {
      console.error('Failed to delete event:', error);
      setIsClosing(false);
    }
  };

  const handleClose = () => {
    setIsClosing(true);
    setTimeout(onClose, 300);
  };

  const getEventTypeIcon = (type: EventType) => {
    return eventTypes.find(t => t.value === type)?.icon || <Tag className="h-4 w-4" />;
  };

  return (
    <>
      <div className={`fixed inset-0 bg-black/50 flex items-center justify-center z-50 
        ${isClosing ? 'animate-fade-out' : 'animate-fade-in'}`}
      >
        <div className={`bg-card rounded-lg shadow-xl w-full max-w-4xl p-6 relative 
          ${isClosing ? 'animate-fade-out' : 'animate-fade-in'}`}
        >
          <button
            onClick={handleClose}
            className="absolute right-4 top-4 text-muted-foreground hover:text-foreground rounded-full p-1 hover:bg-accent/50 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>

          <div className="flex items-center gap-3 mb-6">
            <div className="h-10 w-2 rounded-full" style={{ backgroundColor: formData.color || '#3b82f6' }} />
            <h2 className="text-xl font-semibold text-foreground">
              {task ? 'Update Event' : 'Create New Event'}
            </h2>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Left column */}
              <div className="space-y-5">
                <div className="bg-background/50 p-5 rounded-lg space-y-4">
                  {/* Title field */}
                  <div>
                    <div className="flex items-center gap-2 mb-1.5">
                      <Edit3 className="h-4 w-4 text-muted-foreground" />
                      <label className="text-sm font-medium text-foreground">Title</label>
                    </div>
                    <input
                      type="text"
                      value={formData.title}
                      onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                      className="block w-full rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-3 py-2"
                      required
                      placeholder="Add title"
                    />
                  </div>

                  {/* Description field */}
                  <div>
                    <div className="flex items-center gap-2 mb-1.5">
                      <Bookmark className="h-4 w-4 text-muted-foreground" />
                      <label className="text-sm font-medium text-foreground">Description</label>
                    </div>
                    <textarea
                      value={formData.description}
                      onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                      className="block w-full rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-3 py-2"
                      rows={3}
                      placeholder="Add description"
                    />
                  </div>
                </div>
                
                {/* Time and Date fields */}
                <div className="bg-background/50 p-5 rounded-lg space-y-4">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Clock className="h-4 w-4 text-muted-foreground" />
                      <label className="text-sm font-medium text-foreground">Time & Date</label>
                    </div>
                    <div className="flex items-center">
                      <Checkbox
                        checked={formData.is_all_day}
                        onChange={(e) => setFormData({ ...formData, is_all_day: e.target.checked })}
                        className="mr-2"
                      />
                      <Label htmlFor="is-all-day" className="text-sm text-foreground">All Day</Label>
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-xs font-medium text-muted-foreground block mb-1">Start</label>
                      <DatePicker
                        selected={formData.start_time}
                        onChange={(date: Date) => setFormData({ ...formData, start_time: date })}
                        showTimeSelect={!formData.is_all_day}
                        dateFormat={formData.is_all_day ? "MMMM d, yyyy" : "MMMM d, yyyy h:mm aa"}
                        className="block w-full rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-3 py-2"
                        timeIntervals={15}
                      />
                    </div>

                    <div>
                      <label className="text-xs font-medium text-muted-foreground block mb-1">End</label>
                      <DatePicker
                        selected={formData.end_time}
                        onChange={(date: Date) => setFormData({ ...formData, end_time: date })}
                        showTimeSelect={!formData.is_all_day}
                        dateFormat={formData.is_all_day ? "MMMM d, yyyy" : "MMMM d, yyyy h:mm aa"}
                        className="block w-full rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-3 py-2"
                        minDate={formData.start_time}
                        timeIntervals={15}
                      />
                    </div>
                  </div>
                </div>
                
                {/* Other fields */}
                <div className="bg-background/50 p-5 rounded-lg space-y-4">
                  <div>
                    <div className="flex items-center gap-2 mb-1.5">
                      <MapPin className="h-4 w-4 text-muted-foreground" />
                      <label className="text-sm font-medium text-foreground">Location</label>
                    </div>
                    <input
                      type="text"
                      value={formData.location}
                      onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                      className="block w-full rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-3 py-2"
                      placeholder="Add location"
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <div className="flex items-center gap-2 mb-1.5">
                        <Tag className="h-4 w-4 text-muted-foreground" />
                        <label className="text-sm font-medium text-foreground">Event Type</label>
                      </div>
                      <Select 
                        value={formData.event_type} 
                        onValueChange={(value) => setFormData({ ...formData, event_type: value as EventType })}
                      >
                        <SelectTrigger className="w-full bg-background border-input text-foreground">
                          <SelectValue placeholder="Select type">
                            <div className="flex items-center gap-2">
                              {getEventTypeIcon(formData.event_type)}
                              <span>{eventTypes.find(t => t.value === formData.event_type)?.label}</span>
                            </div>
                          </SelectValue>
                        </SelectTrigger>
                        <SelectContent className="bg-popover border-input text-foreground">
                          <SelectGroup>
                            {eventTypes.map(type => (
                              <SelectItem key={type.value} value={type.value} className="focus:bg-accent">
                                <div className="flex items-center gap-2">
                                  {type.icon}
                                  <span>{type.label}</span>
                                </div>
                              </SelectItem>
                            ))}
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </div>
                    {/* Color picker */}
                  <div>
                    <div className="flex items-center gap-2 mb-1.5">
                      <span 
                        className="inline-block h-4 w-4 rounded-full" 
                        style={{ backgroundColor: formData.color || '#3b82f6' }} 
                      />
                      <label className="text-sm font-medium text-foreground">Color</label>
                    </div>
                    <Popover>
                      <PopoverTrigger asChild>
                        <button
                          type="button"
                          className="flex items-center justify-between w-full rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-3 py-2"
                        >
                          <div className="flex items-center gap-2">
                            <span 
                              className="inline-block h-4 w-4 rounded-full" 
                              style={{ backgroundColor: formData.color || '#3b82f6' }} 
                            />
                            <span>
                              {colorOptions.find(c => c.value === formData.color)?.name || 'Custom'}
                            </span>
                          </div>
                          <ChevronDown className="h-4 w-4 opacity-50" />
                        </button>
                      </PopoverTrigger>
                      <PopoverContent className="w-48 p-2 bg-popover border-input text-foreground">
                        <div className="grid grid-cols-4 gap-1.5 mb-2">
                          {colorOptions.map(color => (
                            <TooltipProvider key={color.value} delayDuration={300}>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <button
                                    type="button"
                                    className={`w-full aspect-square rounded-full border-2 ${
                                      formData.color === color.value ? 'border-primary' : 'border-transparent'
                                    } hover:scale-110 transition-transform`}
                                    style={{ backgroundColor: color.value }}
                                    onClick={() => {
                                      setFormData({ ...formData, color: color.value });
                                      setShowCustomColor(false);
                                    }}
                                  />
                                </TooltipTrigger>
                                <TooltipContent side="bottom" className="py-1 px-2 text-xs">
                                  <p>{color.name}</p>
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          ))}
                        </div>
                        
                        <div className="border-t border-border pt-2">
                          <button
                            type="button"
                            className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground w-full"
                            onClick={() => setShowCustomColor(!showCustomColor)}
                          >
                            <span className="text-xs">+</span>
                            <span>Custom color</span>
                          </button>
                          
                          {showCustomColor && (
                            <div className="mt-1.5">
                              <input
                                type="color"
                                value={formData.color || '#3b82f6'}
                                onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                                className="w-full h-6 rounded cursor-pointer"
                              />
                            </div>
                          )}
                        </div>
                      </PopoverContent>
                    </Popover>
                  </div>              
                  </div>
                  
                  
                </div>
              </div>
              
              {/* Right column */}
              <div className="space-y-5">
                <div className="bg-background/50 p-5 rounded-lg">
                  <div className="flex items-center gap-2 mb-4">
                    <Repeat className="h-4 w-4 text-muted-foreground" />
                    <label className="text-sm font-medium text-foreground">Recurrence</label>
                    <div className="flex-1"></div>
                    <div className="relative inline-flex">
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={recurrenceData.enabled}
                          onChange={(e) => setRecurrenceData(prev => ({ 
                            ...prev, 
                            enabled: e.target.checked,
                            freq: e.target.checked ? 'Daily' : 'None'
                          }))}
                          className="sr-only peer"
                        />
                        <div className="w-11 h-6 bg-muted peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-foreground after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-gray-700"></div>
                      </label>
                    </div>
                  </div>

                  {recurrenceData.enabled && (
                    <div className="space-y-4">
                      <div>
                        <label className="text-xs font-medium text-muted-foreground block mb-1">Frequency</label>
                        <Select 
                          value={recurrenceData.freq} 
                          onValueChange={(value) => setRecurrenceData(prev => ({
                            ...prev,
                            freq: value as RecurrenceType
                          }))}
                        >
                          <SelectTrigger className="w-full bg-background border-input text-foreground">
                            <SelectValue placeholder="Select frequency" />
                          </SelectTrigger>
                          <SelectContent className="bg-popover border-input text-foreground">
                            <SelectGroup>
                              {recurrenceTypes.filter(t => t.value !== 'None').map(type => (
                                <SelectItem key={type.value} value={type.value} className="focus:bg-accent">
                                  {type.label}
                                </SelectItem>
                              ))}
                            </SelectGroup>
                          </SelectContent>
                        </Select>
                      </div>

                      <div>
                        <label className="text-xs font-medium text-muted-foreground block mb-1">Repeat every</label>
                        <div className="flex items-center gap-2">
                          <input
                            type="number"
                            min="1"
                            value={recurrenceData.interval}
                            onChange={(e) => setRecurrenceData(prev => ({
                              ...prev,
                              interval: parseInt(e.target.value) || 1
                            }))}
                            className="block w-20 rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-3 py-2"
                          />
                          <span className="text-sm text-foreground">
                            {recurrenceData.freq.toLowerCase()}
                            {recurrenceData.interval > 1 ? 's' : ''}
                          </span>
                        </div>
                      </div>

                      <Separator className="bg-border" />

                      <div>
                        <label className="text-xs font-medium text-muted-foreground block mb-3">Ends</label>
                        <RadioGroup 
                          value={recurrenceData.count ? "count" : recurrenceData.until ? "until" : "never"}
                          onValueChange={(value) => {
                            switch(value) {
                              case "never":
                                setRecurrenceData(prev => ({
                                  ...prev,
                                  until: null,
                                  count: null
                                }));
                                break;
                              case "count":
                                setRecurrenceData(prev => ({
                                  ...prev,
                                  until: null,
                                  count: 1
                                }));
                                break;
                              case "until":
                                setRecurrenceData(prev => ({
                                  ...prev,
                                  until: new Date(),
                                  count: null
                                }));
                                break;
                            }
                          }}
                          className="space-y-2"
                        >
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="never" id="never" />
                            <Label htmlFor="never" className="text-sm text-foreground">Never</Label>
                          </div>
                          
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="count" id="count" />
                            <Label htmlFor="count" className="text-sm text-foreground">After</Label>
                            {recurrenceData.count !== null && (
                              <input
                                type="number"
                                min="1"
                                value={recurrenceData.count}
                                onChange={(e) => setRecurrenceData(prev => ({
                                  ...prev,
                                  count: parseInt(e.target.value) || 1
                                }))}
                                className="ml-2 w-16 rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-2 py-1 text-sm"
                              />
                            )}
                            <span className="text-sm text-foreground">occurrences</span>
                          </div>
                          
                          <div className="flex items-center space-x-2">
                            <RadioGroupItem value="until" id="until" />
                            <Label htmlFor="until" className="text-sm text-foreground">On date</Label>
                            {recurrenceData.until && (
                              <div className="ml-2 flex-1">
                                <DatePicker
                                  selected={recurrenceData.until}
                                  onChange={(date: Date) => setRecurrenceData(prev => ({
                                    ...prev,
                                    until: date
                                  }))}
                                  minDate={formData.start_time}
                                  dateFormat="MMMM d, yyyy"
                                  className="w-full rounded-md bg-background border border-input text-foreground shadow-sm focus:border-primary focus:ring-primary px-3 py-1 text-sm"
                                />
                              </div>
                            )}
                          </div>
                        </RadioGroup>
                      </div>
                    </div>
                  )}
                </div>
                
                {/* Status section */}
                <div className="bg-background/50 p-5 rounded-lg">
                  <div className="flex items-center gap-2 mb-3">
                    <AlertCircle className="h-4 w-4 text-muted-foreground" />
                    <label className="text-sm font-medium text-foreground">Status</label>
                  </div>
                  
                  <Select 
                    value={formData.status} 
                    onValueChange={(value) => setFormData({ ...formData, status: value as typeof formData.status })}
                  >
                    <SelectTrigger className="w-full bg-background border-input text-foreground">
                      <SelectValue placeholder="Select status" />
                    </SelectTrigger>
                    <SelectContent className="bg-popover border-input text-foreground">
                      <SelectGroup>
                        {statusTypes.map(type => (
                          <SelectItem key={type.value} value={type.value} className="focus:bg-accent">
                            {type.label}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            </div>

            <div className="flex justify-between gap-3 pt-4 border-t border-border">
              {task && (
                <Button
                  type="button"
                  onClick={handleDelete}
                  variant="destructive"
                  className="px-4 py-2"
                  disabled={deleteEvent.isPending}
                >
                  {deleteEvent.isPending ? 'Deleting...' : 'Delete Event'}
                </Button>
              )}
              <div className="flex gap-3 ml-auto">
                <Button
                  type="button"
                  onClick={handleClose}
                  variant="outline"
                  className="px-4 py-2"
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  disabled={createEvent.isPending || updateEvent.isPending}
                  className="px-4 py-2"
                >
                  {createEvent.isPending || updateEvent.isPending 
                    ? 'Saving...' 
                    : task ? 'Update Event' : 'Create Event'}
                </Button>
              </div>
            </div>
          </form>
        </div>
      </div>

      {/* Update Confirmation Dialog */}
      <Dialog open={showUpdateModal} onOpenChange={setShowUpdateModal}>
        <DialogContent className="bg-card text-foreground border-border max-w-md">
          <DialogHeader>
            <DialogTitle>Update recurring event</DialogTitle>
            <DialogDescription className="text-muted-foreground">
              Would you like to update just this occurrence or all events in the series?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="flex flex-col sm:flex-row gap-2 sm:gap-0">
            <Button 
              variant="outline" 
              className="w-full sm:w-auto"
              onClick={() => {
                setShowUpdateModal(false);
                saveEvent('single');
              }}
            >
              Update only this event
            </Button>
            <Button 
              className="w-full sm:w-auto"
              onClick={() => {
                setShowUpdateModal(false);
                saveEvent('all');
              }}
            >
              Update all events
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={showDeleteModal} onOpenChange={setShowDeleteModal}>
        <DialogContent className="bg-card text-foreground border-border max-w-md">
          <DialogHeader>
            <DialogTitle>Delete recurring event</DialogTitle>
            <DialogDescription className="text-muted-foreground">
              Would you like to delete just this occurrence or all events in the series?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="flex flex-col sm:flex-row gap-2 sm:gap-0">
            <Button 
              variant="destructive" 
              className="w-full sm:w-auto"
              onClick={() => {
                setShowDeleteModal(false);
                deleteEventConfirmed('single');
              }}
            >
              Delete only this event
            </Button>
            <Button 
              variant="destructive"
              className="w-full sm:w-auto"
              onClick={() => {
                setShowDeleteModal(false);
                deleteEventConfirmed('all');
              }}
            >
              Delete all events
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default EventForm;