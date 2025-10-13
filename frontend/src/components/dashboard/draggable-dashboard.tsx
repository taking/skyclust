'use client';

import { useState, useEffect } from 'react';
import { DndContext, DragEndEvent, DragOverlay, DragStartEvent, closestCenter } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy, arrayMove } from '@dnd-kit/sortable';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { GripVertical, Settings, X } from 'lucide-react';
import { WidgetData, WidgetType, WIDGET_CONFIGS, getWidgetSizeClasses } from '@/lib/widgets';
import { VMStatusWidget } from '@/components/widgets/vm-status-widget';
import { CostChartWidget } from '@/components/widgets/cost-chart-widget';
import { ResourceUsageWidget } from '@/components/widgets/resource-usage-widget';

interface DraggableDashboardProps {
  widgets: WidgetData[];
  onWidgetsChange: (widgets: WidgetData[]) => void;
  onWidgetRemove: (widgetId: string) => void;
  onWidgetConfigure: (widgetId: string) => void;
}

interface DraggableWidgetProps {
  widget: WidgetData;
  onRemove: (widgetId: string) => void;
  onConfigure: (widgetId: string) => void;
}

function DraggableWidget({ widget, onRemove, onConfigure }: DraggableWidgetProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: widget.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const config = WIDGET_CONFIGS[widget.type];

  const renderWidget = () => {
    switch (widget.type) {
      case 'vm-status':
        return <VMStatusWidget />;
      case 'cost-chart':
        return <CostChartWidget />;
      case 'resource-usage':
        return <ResourceUsageWidget />;
      default:
        return (
          <div className="flex items-center justify-center h-32 text-gray-500">
            Widget not implemented
          </div>
        );
    }
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`${getWidgetSizeClasses(widget.size)} ${
        isDragging ? 'opacity-50' : ''
      }`}
    >
      <Card className="h-full">
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div
                {...attributes}
                {...listeners}
                className="mr-2 cursor-grab active:cursor-grabbing"
              >
                <GripVertical className="h-4 w-4 text-gray-400" />
              </div>
              <div>
                <CardTitle className="text-sm">{config.title}</CardTitle>
                <CardDescription className="text-xs">{config.description}</CardDescription>
              </div>
            </div>
            <div className="flex items-center space-x-1">
              <Badge variant="outline" className="text-xs">
                {widget.size}
              </Badge>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onConfigure(widget.id)}
                className="h-6 w-6 p-0"
              >
                <Settings className="h-3 w-3" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRemove(widget.id)}
                className="h-6 w-6 p-0 text-red-500 hover:text-red-700"
              >
                <X className="h-3 w-3" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="pt-0">
          {renderWidget()}
        </CardContent>
      </Card>
    </div>
  );
}

export function DraggableDashboard({
  widgets,
  onWidgetsChange,
  onWidgetRemove,
  onWidgetConfigure,
}: DraggableDashboardProps) {
  const [activeId, setActiveId] = useState<string | null>(null);

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    if (active.id !== over?.id) {
      const oldIndex = widgets.findIndex((widget) => widget.id === active.id);
      const newIndex = widgets.findIndex((widget) => widget.id === over?.id);

      const newWidgets = arrayMove(widgets, oldIndex, newIndex);
      onWidgetsChange(newWidgets);
    }

    setActiveId(null);
  };

  const activeWidget = widgets.find((widget) => widget.id === activeId);

  return (
    <DndContext
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        <SortableContext items={widgets.map(w => w.id)} strategy={verticalListSortingStrategy}>
          {widgets.map((widget) => (
            <DraggableWidget
              key={widget.id}
              widget={widget}
              onRemove={onWidgetRemove}
              onConfigure={onWidgetConfigure}
            />
          ))}
        </SortableContext>
      </div>

      <DragOverlay>
        {activeWidget ? (
          <div className={`${getWidgetSizeClasses(activeWidget.size)} opacity-90`}>
            <Card className="h-full shadow-lg">
              <CardHeader className="pb-3">
                <div className="flex items-center">
                  <GripVertical className="mr-2 h-4 w-4 text-gray-400" />
                  <div>
                    <CardTitle className="text-sm">{WIDGET_CONFIGS[activeWidget.type].title}</CardTitle>
                    <CardDescription className="text-xs">{WIDGET_CONFIGS[activeWidget.type].description}</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="flex items-center justify-center h-32 text-gray-500">
                  Dragging...
                </div>
              </CardContent>
            </Card>
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
