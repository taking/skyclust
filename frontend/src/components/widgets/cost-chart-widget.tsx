'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { DollarSign, TrendingUp, TrendingDown } from 'lucide-react';

interface CostChartWidgetProps {
  data?: {
    monthly: Array<{ month: string; cost: number }>;
    breakdown: Array<{ service: string; cost: number; percentage: number }>;
    total: number;
    change: number;
  };
  isLoading?: boolean;
}

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8'];

export function CostChartWidget({ data, isLoading }: CostChartWidgetProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center">
            <DollarSign className="mr-2 h-5 w-5" />
            Cost Analysis
          </CardTitle>
          <CardDescription>Loading cost data...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-64 animate-pulse bg-gray-200 rounded"></div>
        </CardContent>
      </Card>
    );
  }

  const mockData = data || {
    monthly: [
      { month: 'Jan', cost: 1200 },
      { month: 'Feb', cost: 1350 },
      { month: 'Mar', cost: 1100 },
      { month: 'Apr', cost: 1400 },
      { month: 'May', cost: 1600 },
      { month: 'Jun', cost: 1800 },
    ],
    breakdown: [
      { service: 'EC2', cost: 800, percentage: 44 },
      { service: 'S3', cost: 300, percentage: 17 },
      { service: 'RDS', cost: 400, percentage: 22 },
      { service: 'Lambda', cost: 200, percentage: 11 },
      { service: 'Other', cost: 100, percentage: 6 },
    ],
    total: 1800,
    change: 12.5,
  };

  const isPositiveChange = mockData.change > 0;

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <DollarSign className="mr-2 h-5 w-5" />
          Cost Analysis
        </CardTitle>
        <CardDescription>
          Monthly spending: ${mockData.total.toLocaleString()}
          <span className={`ml-2 flex items-center ${isPositiveChange ? 'text-red-600' : 'text-green-600'}`}>
            {isPositiveChange ? (
              <TrendingUp className="mr-1 h-3 w-3" />
            ) : (
              <TrendingDown className="mr-1 h-3 w-3" />
            )}
            {Math.abs(mockData.change)}%
          </span>
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* Monthly Trend Chart */}
          <div>
            <h4 className="text-sm font-medium mb-2">Monthly Trend</h4>
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={mockData.monthly}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="month" />
                <YAxis />
                <Tooltip formatter={(value) => [`$${value}`, 'Cost']} />
                <Line 
                  type="monotone" 
                  dataKey="cost" 
                  stroke="#8884d8" 
                  strokeWidth={2}
                  dot={{ fill: '#8884d8', strokeWidth: 2, r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Service Breakdown */}
          <div>
            <h4 className="text-sm font-medium mb-2">Service Breakdown</h4>
            <div className="flex items-center space-x-4">
              <div className="w-32 h-32">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={mockData.breakdown}
                      cx="50%"
                      cy="50%"
                      innerRadius={30}
                      outerRadius={60}
                      paddingAngle={2}
                      dataKey="cost"
                    >
                      {mockData.breakdown.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value) => [`$${value}`, 'Cost']} />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="flex-1 space-y-2">
                {mockData.breakdown.map((item, index) => (
                  <div key={item.service} className="flex items-center justify-between">
                    <div className="flex items-center">
                      <div 
                        className="w-3 h-3 rounded-full mr-2" 
                        style={{ backgroundColor: COLORS[index % COLORS.length] }}
                      />
                      <span className="text-sm">{item.service}</span>
                    </div>
                    <div className="text-right">
                      <div className="text-sm font-medium">${item.cost}</div>
                      <div className="text-xs text-gray-500">{item.percentage}%</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
