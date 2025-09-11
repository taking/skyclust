import React from 'react'
import {
  Card,
  Text,
  Button,
  Group,
  Stack,
  Badge,
  Grid,
  Title,
  Progress,
} from '@mantine/core'
import { IconCode, IconPlay, IconCheck, IconX } from '@tabler/icons-react'

export default function TofuExecutions() {
  // Mock data for now
  const executions = [
    {
      id: 'exec-1',
      command: 'plan',
      status: 'completed',
      started_at: '2024-01-15T10:30:00Z',
      completed_at: '2024-01-15T10:32:00Z',
    },
    {
      id: 'exec-2',
      command: 'apply',
      status: 'running',
      started_at: '2024-01-15T11:00:00Z',
      completed_at: null,
    },
    {
      id: 'exec-3',
      command: 'destroy',
      status: 'failed',
      started_at: '2024-01-15T09:15:00Z',
      completed_at: '2024-01-15T09:17:00Z',
    },
  ]

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'green'
      case 'running':
        return 'blue'
      case 'failed':
        return 'red'
      case 'pending':
        return 'yellow'
      default:
        return 'gray'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <IconCheck size={16} />
      case 'running':
        return <IconPlay size={16} />
      case 'failed':
        return <IconX size={16} />
      default:
        return <IconCode size={16} />
    }
  }

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>OpenTofu Executions</Title>
        <Group>
          <Button leftSection={<IconCode size={16} />} variant="outline">
            Plan
          </Button>
          <Button leftSection={<IconPlay size={16} />}>
            Apply
          </Button>
        </Group>
      </Group>

      {executions.length === 0 ? (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Stack align="center" py="xl">
            <IconCode size={48} color="gray" />
            <Text size="lg" c="dimmed">No executions found</Text>
            <Text size="sm" c="dimmed">
              Run your first OpenTofu command
            </Text>
          </Stack>
        </Card>
      ) : (
        <Grid>
          {executions.map((exec) => (
            <Grid.Col key={exec.id} span={{ base: 12, sm: 6, md: 4 }}>
              <Card shadow="sm" padding="lg" radius="md" withBorder>
                <Stack>
                  <Group justify="space-between">
                    <Text fw={500} size="lg">
                      {exec.command}
                    </Text>
                    <Badge
                      color={getStatusColor(exec.status)}
                      leftSection={getStatusIcon(exec.status)}
                    >
                      {exec.status}
                    </Badge>
                  </Group>
                  
                  <Text size="sm" c="dimmed">
                    Started: {new Date(exec.started_at).toLocaleString()}
                  </Text>
                  
                  {exec.completed_at && (
                    <Text size="sm" c="dimmed">
                      Completed: {new Date(exec.completed_at).toLocaleString()}
                    </Text>
                  )}
                  
                  {exec.status === 'running' && (
                    <Progress value={65} size="sm" />
                  )}
                </Stack>
              </Card>
            </Grid.Col>
          ))}
        </Grid>
      )}
    </Stack>
  )
}

