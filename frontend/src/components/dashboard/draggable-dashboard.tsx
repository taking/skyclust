'use client';

import * as React from 'react';
import { useState, useEffect } from 'react';
import dynamic from 'next/dynamic';
import { DndContext, DragEndEvent, DragOverlay, DragStartEvent, closestCenter } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy, arrayMove } from '@dnd-kit/sortable';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { GripVertical, Settings, X, Maximize2, Minimize2 } from 'lucide-react';
import { WidgetData, WidgetType, WidgetSize, WIDGET_CONFIGS, getWidgetSizeClasses, useWidgetConfigs } from '@/lib/widgets';
import { useTranslation } from '@/hooks/use-translation';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

// Dynamic imports for widgets with loading states
const VMStatusWidget = dynamic(
  () => import('@/components/widgets/vm-status-widget').then(mod => ({ default: mod.VMStatusWidget })),
  { 
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    ),
  }
);

const CostChartWidget = dynamic(
  () => import('@/components/widgets/cost-chart-widget').then(mod => ({ default: mod.CostChartWidget })),
  { 
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    ),
  }
);

const ResourceUsageWidget = dynamic(
  () => import('@/components/widgets/resource-usage-widget').then(mod => ({ default: mod.ResourceUsageWidget })),
  { 
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    ),
  }
);

const KubernetesStatusWidget = dynamic(
  () => import('@/components/widgets/kubernetes-status-widget').then(mod => ({ default: mod.KubernetesStatusWidget })),
  { 
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    ),
  }
);

const NetworkStatusWidget = dynamic(
  () => import('@/components/widgets/network-status-widget').then(mod => ({ default: mod.NetworkStatusWidget })),
  { 
    ssr: false,
    loading: () => (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    ),
  }
);

interface DraggableDashboardProps {
  widgets: WidgetData[];
  onWidgetsChange: (widgets: WidgetData[]) => void;
  onWidgetRemove: (widgetId: string) => void;
  onWidgetConfigure: (widgetId: string) => void;
  onWidgetResize?: (widgetId: string, size: WidgetSize) => void;
}

interface DraggableWidgetProps {
  widget: WidgetData;
  onRemove: (widgetId: string) => void;
  onConfigure: (widgetId: string) => void;
  onResize?: (widgetId: string, size: WidgetSize) => void;
}

const DraggableWidget = React.memo(function DraggableWidget({ widget, onRemove, onConfigure, onResize }: DraggableWidgetProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: widget.id });
  const { t } = useTranslation();
  const { getWidgetConfig } = useWidgetConfigs();

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const config = getWidgetConfig(widget.type);
  
  const sizeOptions: WidgetSize[] = ['small', 'medium', 'large', 'xlarge'];
  const availableSizes = sizeOptions.filter(s => {
    const index = sizeOptions.indexOf(s);
    const minIndex = sizeOptions.indexOf(config.minSize);
    const maxIndex = sizeOptions.indexOf(config.maxSize);
    return index >= minIndex && index <= maxIndex;
  });

  const handleSizeChange = (newSize: WidgetSize) => {
    if (onResize) {
      onResize(widget.id, newSize);
    }
  };

  const renderWidget = () => {
    switch (widget.type) {
      case 'vm-status':
        return <VMStatusWidget />;
      case 'cost-chart':
        return <CostChartWidget />;
      case 'resource-usage':
        return <ResourceUsageWidget />;
      case 'kubernetes-status':
        return <KubernetesStatusWidget />;
      case 'network-status':
        return <NetworkStatusWidget />;
      default:
        return (
          <div className="flex items-center justify-center h-32 text-gray-500">
            {t('widgets.addWidget.notImplemented')}
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
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-6 w-6 p-0"
                    aria-label="Resize widget"
                  >
                    {widget.size === 'xlarge' || widget.size === 'large' ? (
                      <Minimize2 className="h-3 w-3" />
                    ) : (
                      <Maximize2 className="h-3 w-3" />
                    )}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {availableSizes.map((size) => (
                    <DropdownMenuItem
                      key={size}
                      onClick={() => handleSizeChange(size)}
                      className={widget.size === size ? 'bg-accent' : ''}
                    >
                      {size.charAt(0).toUpperCase() + size.slice(1)}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onConfigure(widget.id)}
                className="h-6 w-6 p-0"
                aria-label="Configure widget"
              >
                <Settings className="h-3 w-3" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRemove(widget.id)}
                className="h-6 w-6 p-0 text-red-500 hover:text-red-700"
                aria-label="Remove widget"
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
});

export function DraggableDashboard({
  widgets,
  onWidgetsChange,
  onWidgetRemove,
  onWidgetConfigure,
  onWidgetResize,
}: DraggableDashboardProps) {
  const { t } = useTranslation();
  const { getWidgetConfig } = useWidgetConfigs();
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
              onResize={onWidgetResize}
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
                    <CardTitle className="text-sm">{getWidgetConfig(activeWidget.type).title}</CardTitle>
                    <CardDescription className="text-xs">{getWidgetConfig(activeWidget.type).description}</CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="flex items-center justify-center h-32 text-gray-500">
                  {t('widgets.addWidget.dragging')}
                </div>
              </CardContent>
            </Card>
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
