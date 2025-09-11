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
  ActionIcon,
} from '@mantine/core'
import { IconServer, IconPlay, IconStop, IconTrash } from '@tabler/icons-react'

export default function VMList() {
  // Mock data for now
  const vms = [
    {
      id: 'vm-1',
      name: 'web-server-01',
      status: 'running',
      type: 't3.micro',
      region: 'us-east-1',
      provider: 'aws',
      created_at: '2024-01-15T10:30:00Z',
    },
    {
      id: 'vm-2',
      name: 'db-server-01',
      status: 'stopped',
      type: 't3.small',
      region: 'us-west-2',
      provider: 'aws',
      created_at: '2024-01-10T14:20:00Z',
    },
  ]

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
        return 'green'
      case 'stopped':
        return 'red'
      case 'pending':
        return 'yellow'
      default:
        return 'gray'
    }
  }

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>Virtual Machines</Title>
        <Button leftSection={<IconServer size={16} />}>
          Create VM
        </Button>
      </Group>

      {vms.length === 0 ? (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Stack align="center" py="xl">
            <IconServer size={48} color="gray" />
            <Text size="lg" c="dimmed">No VMs found</Text>
            <Text size="sm" c="dimmed">
              Create your first virtual machine
            </Text>
          </Stack>
        </Card>
      ) : (
        <Grid>
          {vms.map((vm) => (
            <Grid.Col key={vm.id} span={{ base: 12, sm: 6, md: 4 }}>
              <Card shadow="sm" padding="lg" radius="md" withBorder>
                <Stack>
                  <Group justify="space-between">
                    <Text fw={500} size="lg">
                      {vm.name}
                    </Text>
                    <Badge color={getStatusColor(vm.status)}>
                      {vm.status}
                    </Badge>
                  </Group>
                  
                  <Group>
                    <Text size="sm" c="dimmed">
                      {vm.type}
                    </Text>
                    <Text size="sm" c="dimmed">
                      {vm.region}
                    </Text>
                  </Group>
                  
                  <Text size="sm" c="dimmed">
                    Provider: {vm.provider}
                  </Text>
                  
                  <Text size="sm" c="dimmed">
                    Created: {new Date(vm.created_at).toLocaleDateString()}
                  </Text>
                  
                  <Group>
                    <ActionIcon variant="subtle" color="green">
                      <IconPlay size={16} />
                    </ActionIcon>
                    <ActionIcon variant="subtle" color="red">
                      <IconStop size={16} />
                    </ActionIcon>
                    <ActionIcon variant="subtle" color="red">
                      <IconTrash size={16} />
                    </ActionIcon>
                  </Group>
                </Stack>
              </Card>
            </Grid.Col>
          ))}
        </Grid>
      )}
    </Stack>
  )
}

