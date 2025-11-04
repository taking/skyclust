'use client';

import { useMemo } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { VPC, Subnet, SecurityGroup } from '@/lib/types';
import { Network, Layers, Shield, ArrowRight } from 'lucide-react';

interface NetworkTopologyViewerProps {
  vpcs: VPC[];
  subnets: Subnet[];
  securityGroups: SecurityGroup[];
  selectedVPCId?: string;
  onVPCClick?: (vpcId: string) => void;
}

interface TopologyNode {
  id: string;
  type: 'vpc' | 'subnet' | 'security-group';
  name: string;
  level: number;
  x: number;
  y: number;
  data: VPC | Subnet | SecurityGroup;
}

export function NetworkTopologyViewer({
  vpcs,
  subnets,
  securityGroups,
  selectedVPCId,
  onVPCClick,
}: NetworkTopologyViewerProps) {
  const topology = useMemo(() => {
    const nodes: TopologyNode[] = [];
    const selectedVPC = vpcs.find(v => v.id === selectedVPCId || selectedVPCId === undefined);
    const activeVPC = selectedVPC || vpcs[0];

    if (!activeVPC) {
      return { nodes: [], connections: [] };
    }

    // VPC Node (Level 0)
    nodes.push({
      id: activeVPC.id,
      type: 'vpc',
      name: activeVPC.name,
      level: 0,
      x: 50,
      y: 50,
      data: activeVPC,
    });

    // Subnets (Level 1)
    const vpcSubnets = subnets.filter(s => s.vpc_id === activeVPC.id);
    vpcSubnets.forEach((subnet, index) => {
      const angle = (index / vpcSubnets.length) * 2 * Math.PI;
      const radius = 30;
      nodes.push({
        id: subnet.id,
        type: 'subnet',
        name: subnet.name,
        level: 1,
        x: 50 + radius * Math.cos(angle),
        y: 50 + radius * Math.sin(angle),
        data: subnet,
      });
    });

    // Security Groups (Level 2)
    const vpcSecurityGroups = securityGroups.filter(sg => sg.vpc_id === activeVPC.id);
    vpcSecurityGroups.forEach((sg, index) => {
      const angle = (index / vpcSecurityGroups.length) * 2 * Math.PI;
      const radius = 45;
      nodes.push({
        id: sg.id,
        type: 'security-group',
        name: sg.name,
        level: 2,
        x: 50 + radius * Math.cos(angle),
        y: 50 + radius * Math.sin(angle),
        data: sg,
      });
    });

    // Connections
    const connections: Array<{ from: string; to: string }> = [];
    
    // VPC to Subnets
    vpcSubnets.forEach(subnet => {
      connections.push({ from: activeVPC.id, to: subnet.id });
    });

    // Subnets to Security Groups (if they have any association)
    vpcSubnets.forEach(subnet => {
      vpcSecurityGroups.forEach(sg => {
        connections.push({ from: subnet.id, to: sg.id });
      });
    });

    return { nodes, connections };
  }, [vpcs, subnets, securityGroups, selectedVPCId]);

  if (vpcs.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Network className="h-12 w-12 text-gray-400 mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">No Network Topology</h3>
          <p className="text-sm text-gray-500 text-center">
            No VPCs found to display topology
          </p>
        </CardContent>
      </Card>
    );
  }

  const getNodeIcon = (type: string) => {
    switch (type) {
      case 'vpc':
        return <Network className="h-5 w-5" />;
      case 'subnet':
        return <Layers className="h-4 w-4" />;
      case 'security-group':
        return <Shield className="h-4 w-4" />;
      default:
        return null;
    }
  };

  const getNodeColor = (type: string) => {
    switch (type) {
      case 'vpc':
        return 'bg-blue-100 text-blue-800 border-blue-300';
      case 'subnet':
        return 'bg-green-100 text-green-800 border-green-300';
      case 'security-group':
        return 'bg-orange-100 text-orange-800 border-orange-300';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-300';
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Network Topology</CardTitle>
        <CardDescription>Visual representation of network resources</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="relative w-full h-96 border border-gray-200 rounded-lg bg-gray-50 overflow-hidden">
          <svg
            viewBox="0 0 100 100"
            className="w-full h-full"
            preserveAspectRatio="xMidYMid meet"
          >
            {/* Connections */}
            {topology.connections.map((conn, index) => {
              const fromNode = topology.nodes.find(n => n.id === conn.from);
              const toNode = topology.nodes.find(n => n.id === conn.to);
              
              if (!fromNode || !toNode) return null;

              return (
                <line
                  key={`${conn.from}-${conn.to}-${index}`}
                  x1={fromNode.x}
                  y1={fromNode.y}
                  x2={toNode.x}
                  y2={toNode.y}
                  stroke="#cbd5e1"
                  strokeWidth="0.3"
                  strokeDasharray="0.5,0.5"
                />
              );
            })}

            {/* Nodes */}
            {topology.nodes.map((node) => (
              <g key={node.id}>
                <circle
                  cx={node.x}
                  cy={node.y}
                  r={node.level === 0 ? 3 : node.level === 1 ? 2 : 1.5}
                  className={getNodeColor(node.type)}
                  stroke="currentColor"
                  strokeWidth="0.2"
                />
                {node.level === 0 && (
                  <text
                    x={node.x}
                    y={node.y - 4}
                    textAnchor="middle"
                    fontSize="2"
                    fill="currentColor"
                    className="font-medium"
                  >
                    {node.name}
                  </text>
                )}
              </g>
            ))}
          </svg>

          {/* Legend */}
          <div className="absolute bottom-4 left-4 bg-white p-3 rounded-md shadow-md border border-gray-200">
            <div className="text-xs font-medium mb-2">Legend</div>
            <div className="space-y-1 text-xs">
              <div className="flex items-center space-x-2">
                <div className={`w-3 h-3 rounded-full ${getNodeColor('vpc')}`} />
                <span>VPC</span>
              </div>
              <div className="flex items-center space-x-2">
                <div className={`w-2 h-2 rounded-full ${getNodeColor('subnet')}`} />
                <span>Subnet</span>
              </div>
              <div className="flex items-center space-x-2">
                <div className={`w-2 h-2 rounded-full ${getNodeColor('security-group')}`} />
                <span>Security Group</span>
              </div>
            </div>
          </div>

          {/* Node Details */}
          <div className="absolute top-4 right-4 bg-white p-4 rounded-md shadow-md border border-gray-200 max-w-xs">
            <div className="text-xs font-medium mb-3">Resource Details</div>
            <div className="space-y-3 text-xs">
              {topology.nodes.map((node) => {
                if (node.type === 'vpc') {
                  const vpc = node.data as VPC;
                  return (
                    <div key={node.id} className="border-b pb-2 last:border-0">
                      <div className="flex items-center space-x-2 mb-1">
                        {getNodeIcon(node.type)}
                        <span className="font-medium">{vpc.name}</span>
                      </div>
                      <div className="pl-6 space-y-1 text-gray-600">
                        <div>State: <Badge variant="outline" className="text-xs">{vpc.state}</Badge></div>
                        {vpc.is_default && <Badge variant="secondary" className="text-xs">Default VPC</Badge>}
                      </div>
                    </div>
                  );
                }
                if (node.type === 'subnet') {
                  const subnet = node.data as Subnet;
                  return (
                    <div key={node.id} className="border-b pb-2 last:border-0">
                      <div className="flex items-center space-x-2 mb-1">
                        {getNodeIcon(node.type)}
                        <span className="font-medium">{subnet.name}</span>
                      </div>
                      <div className="pl-6 space-y-1 text-gray-600">
                        <div>CIDR: {subnet.cidr_block}</div>
                        <div>AZ: {subnet.availability_zone}</div>
                      </div>
                    </div>
                  );
                }
                if (node.type === 'security-group') {
                  const sg = node.data as SecurityGroup;
                  return (
                    <div key={node.id} className="border-b pb-2 last:border-0">
                      <div className="flex items-center space-x-2 mb-1">
                        {getNodeIcon(node.type)}
                        <span className="font-medium">{sg.name}</span>
                      </div>
                      <div className="pl-6 space-y-1 text-gray-600">
                        <div>Rules: {sg.rules?.length || 0}</div>
                      </div>
                    </div>
                  );
                }
                return null;
              })}
            </div>
          </div>
        </div>

        {/* VPC Selector */}
        {vpcs.length > 1 && (
          <div className="mt-4">
            <div className="text-sm font-medium mb-2">Select VPC</div>
            <div className="flex flex-wrap gap-2">
              {vpcs.map((vpc) => (
                <button
                  key={vpc.id}
                  onClick={() => onVPCClick?.(vpc.id)}
                  className={`px-3 py-1 rounded-md text-sm border transition-colors ${
                    (selectedVPCId || vpcs[0]?.id) === vpc.id
                      ? 'bg-blue-100 border-blue-500 text-blue-900'
                      : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                  }`}
                >
                  {vpc.name}
                </button>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

