'use client';

import { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Plus, Search, Server, DollarSign, Activity, Clock, Zap, AlertTriangle, TrendingUp, MapPin } from 'lucide-react';
import { WidgetType, WIDGET_CONFIGS } from '@/lib/widgets';

interface WidgetAddPanelProps {
  onAddWidget: (type: WidgetType) => void;
  existingWidgets: string[];
}

const iconMap = {
  Server,
  DollarSign,
  Activity,
  Clock,
  Zap,
  AlertTriangle,
  TrendingUp,
  MapPin,
};

export function WidgetAddPanel({ onAddWidget, existingWidgets }: WidgetAddPanelProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  const categories = [
    { id: 'all', label: 'All', count: Object.keys(WIDGET_CONFIGS).length },
    { id: 'overview', label: 'Overview', count: Object.values(WIDGET_CONFIGS).filter(w => w.category === 'overview').length },
    { id: 'monitoring', label: 'Monitoring', count: Object.values(WIDGET_CONFIGS).filter(w => w.category === 'monitoring').length },
    { id: 'cost', label: 'Cost', count: Object.values(WIDGET_CONFIGS).filter(w => w.category === 'cost').length },
    { id: 'management', label: 'Management', count: Object.values(WIDGET_CONFIGS).filter(w => w.category === 'management').length },
  ];

  const filteredWidgets = Object.entries(WIDGET_CONFIGS).filter(([type, config]) => {
    const matchesSearch = config.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         config.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = selectedCategory === 'all' || config.category === selectedCategory;
    const notAlreadyAdded = !existingWidgets.includes(type);
    
    return matchesSearch && matchesCategory && notAlreadyAdded;
  });

  const handleAddWidget = (type: WidgetType) => {
    onAddWidget(type);
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Add Widget
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle>Add Widget</DialogTitle>
          <DialogDescription>
            Choose a widget to add to your dashboard
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Search and Filter */}
          <div className="flex space-x-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="Search widgets..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
          </div>

          {/* Category Tabs */}
          <Tabs value={selectedCategory} onValueChange={setSelectedCategory}>
            <TabsList className="grid w-full grid-cols-5">
              {categories.map((category) => (
                <TabsTrigger key={category.id} value={category.id}>
                  {category.label} ({category.count})
                </TabsTrigger>
              ))}
            </TabsList>

            <TabsContent value={selectedCategory} className="mt-4">
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-h-96 overflow-y-auto">
                {filteredWidgets.map(([type, config]) => {
                  const IconComponent = iconMap[config.icon as keyof typeof iconMap] || Server;
                  
                  return (
                    <Card key={type} className="cursor-pointer hover:shadow-md transition-shadow">
                      <CardHeader className="pb-3">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center">
                            <IconComponent className="mr-2 h-5 w-5 text-blue-500" />
                            <CardTitle className="text-sm">{config.title}</CardTitle>
                          </div>
                          <Badge variant="outline" className="text-xs">
                            {config.defaultSize}
                          </Badge>
                        </div>
                        <CardDescription className="text-xs">
                          {config.description}
                        </CardDescription>
                      </CardHeader>
                      <CardContent className="pt-0">
                        <div className="space-y-2">
                          <div className="flex items-center justify-between text-xs text-gray-500">
                            <span>Size: {config.defaultSize}</span>
                            <span>Category: {config.category}</span>
                          </div>
                          <Button
                            size="sm"
                            className="w-full"
                            onClick={() => handleAddWidget(type as WidgetType)}
                          >
                            <Plus className="mr-1 h-3 w-3" />
                            Add Widget
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  );
                })}
              </div>

              {filteredWidgets.length === 0 && (
                <div className="text-center py-8">
                  <div className="mx-auto h-12 w-12 text-gray-400">
                    <Search className="h-12 w-12" />
                  </div>
                  <h3 className="mt-2 text-sm font-medium text-gray-900">No widgets found</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    {searchQuery 
                      ? 'Try adjusting your search criteria.'
                      : 'All available widgets have been added to your dashboard.'
                    }
                  </p>
                </div>
              )}
            </TabsContent>
          </Tabs>
        </div>
      </DialogContent>
    </Dialog>
  );
}
