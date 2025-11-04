/**
 * Subnet Table Component
 * Subnet 목록 테이블 컴포넌트
 */

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Pagination } from '@/components/ui/pagination';
import { SearchBar } from '@/components/ui/search-bar';
import { Trash2, Edit } from 'lucide-react';
import { UI } from '@/lib/constants';
import type { Subnet } from '@/lib/types';

interface SubnetTableProps {
  subnets: Subnet[];
  filteredSubnets: Subnet[];
  paginatedSubnets: Subnet[];
  selectedSubnetIds: string[];
  onSelectionChange: (ids: string[]) => void;
  onDelete: (subnetId: string, region: string) => void;
  searchQuery: string;
  onSearchChange: (query: string) => void;
  onSearchClear: () => void;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  isDeleting?: boolean;
}

export function SubnetTable({
  subnets,
  filteredSubnets,
  paginatedSubnets,
  selectedSubnetIds,
  onSelectionChange,
  onDelete,
  searchQuery,
  onSearchChange,
  onSearchClear,
  page,
  pageSize,
  onPageChange,
  onPageSizeChange,
  isDeleting = false,
}: SubnetTableProps) {
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      onSelectionChange(filteredSubnets.map(s => s.id));
    } else {
      onSelectionChange([]);
    }
  };

  const handleSelectOne = (subnetId: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedSubnetIds, subnetId]);
    } else {
      onSelectionChange(selectedSubnetIds.filter(id => id !== subnetId));
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Subnets</CardTitle>
        <CardDescription>
          {filteredSubnets.length} of {subnets.length} subnet{subnets.length !== 1 ? 's' : ''} found
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="mb-4">
          <SearchBar
            value={searchQuery}
            onChange={onSearchChange}
            onClear={onSearchClear}
            placeholder="Search subnets by name, CIDR, or state..."
          />
        </div>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={selectedSubnetIds.length === filteredSubnets.length && filteredSubnets.length > 0}
                  onCheckedChange={handleSelectAll}
                />
              </TableHead>
              <TableHead>Name</TableHead>
              <TableHead>CIDR Block</TableHead>
              <TableHead>Availability Zone</TableHead>
              <TableHead>State</TableHead>
              <TableHead>Public</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedSubnets.map((subnet) => {
              const isSelected = selectedSubnetIds.includes(subnet.id);
              
              return (
                <TableRow key={subnet.id}>
                  <TableCell>
                    <Checkbox
                      checked={isSelected}
                      onCheckedChange={(checked) => handleSelectOne(subnet.id, checked)}
                    />
                  </TableCell>
                  <TableCell className="font-medium">{subnet.name}</TableCell>
                  <TableCell>{subnet.cidr_block}</TableCell>
                  <TableCell>{subnet.availability_zone}</TableCell>
                  <TableCell>
                    <Badge variant={subnet.state === 'available' ? 'default' : 'secondary'}>
                      {subnet.state}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    {subnet.is_public ? (
                      <Badge variant="outline">Public</Badge>
                    ) : (
                      <Badge variant="secondary">Private</Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                      <Button variant="ghost" size="sm">
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDelete(subnet.id, subnet.region)}
                        disabled={isDeleting}
                      >
                        <Trash2 className="h-4 w-4 text-red-600" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
        
        {filteredSubnets.length > 0 && (
          <div className="border-t mt-4">
            <Pagination
              total={filteredSubnets.length}
              page={page}
              pageSize={pageSize}
              onPageChange={onPageChange}
              onPageSizeChange={onPageSizeChange}
              pageSizeOptions={UI.PAGINATION.PAGE_SIZE_OPTIONS}
              showPageSizeSelector={true}
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

