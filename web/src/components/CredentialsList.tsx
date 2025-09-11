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
import { IconKey, IconEdit, IconTrash, IconPlus } from '@tabler/icons-react'

export default function CredentialsList() {
  // Mock data for now
  const credentials = [
    {
      id: 'cred-1',
      provider: 'aws',
      name: 'AWS Production',
      created_at: '2024-01-15T10:30:00Z',
    },
    {
      id: 'cred-2',
      provider: 'gcp',
      name: 'GCP Development',
      created_at: '2024-01-10T14:20:00Z',
    },
  ]

  const getProviderColor = (provider: string) => {
    switch (provider) {
      case 'aws':
        return 'orange'
      case 'gcp':
        return 'blue'
      case 'azure':
        return 'cyan'
      default:
        return 'gray'
    }
  }

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>Cloud Credentials</Title>
        <Button leftSection={<IconPlus size={16} />}>
          Add Credentials
        </Button>
      </Group>

      {credentials.length === 0 ? (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Stack align="center" py="xl">
            <IconKey size={48} color="gray" />
            <Text size="lg" c="dimmed">No credentials found</Text>
            <Text size="sm" c="dimmed">
              Add your first cloud provider credentials
            </Text>
          </Stack>
        </Card>
      ) : (
        <Grid>
          {credentials.map((cred) => (
            <Grid.Col key={cred.id} span={{ base: 12, sm: 6, md: 4 }}>
              <Card shadow="sm" padding="lg" radius="md" withBorder>
                <Stack>
                  <Group justify="space-between">
                    <Text fw={500} size="lg">
                      {cred.name}
                    </Text>
                    <Badge color={getProviderColor(cred.provider)}>
                      {cred.provider.toUpperCase()}
                    </Badge>
                  </Group>
                  
                  <Text size="sm" c="dimmed">
                    Created: {new Date(cred.created_at).toLocaleDateString()}
                  </Text>
                  
                  <Group>
                    <ActionIcon variant="subtle" color="blue">
                      <IconEdit size={16} />
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

