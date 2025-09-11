import React, { useState } from 'react'
import {
  Card,
  Text,
  Button,
  Group,
  Stack,
  TextInput,
  Modal,
  ActionIcon,
  Badge,
  Grid,
  Title,
} from '@mantine/core'
import { IconPlus, IconEdit, IconTrash, IconUsers } from '@tabler/icons-react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notifications } from '@mantine/notifications'
import { workspacesApi, Workspace, CreateWorkspaceRequest } from '../api/workspaces'

export default function WorkspaceList() {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [newWorkspaceName, setNewWorkspaceName] = useState('')
  const queryClient = useQueryClient()

  // Fetch workspaces
  const { data: workspacesData, isLoading } = useQuery({
    queryKey: ['workspaces'],
    queryFn: workspacesApi.list,
  })

  // Create workspace mutation
  const createMutation = useMutation({
    mutationFn: workspacesApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workspaces'] })
      setIsCreateModalOpen(false)
      setNewWorkspaceName('')
      notifications.show({
        title: 'Success',
        message: 'Workspace created successfully',
        color: 'green',
      })
    },
    onError: (error: any) => {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.error || 'Failed to create workspace',
        color: 'red',
      })
    },
  })

  // Delete workspace mutation
  const deleteMutation = useMutation({
    mutationFn: workspacesApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workspaces'] })
      notifications.show({
        title: 'Success',
        message: 'Workspace deleted successfully',
        color: 'green',
      })
    },
    onError: (error: any) => {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.error || 'Failed to delete workspace',
        color: 'red',
      })
    },
  })

  const handleCreateWorkspace = () => {
    if (newWorkspaceName.trim()) {
      createMutation.mutate({ name: newWorkspaceName.trim() })
    }
  }

  const handleDeleteWorkspace = (id: string) => {
    if (window.confirm('Are you sure you want to delete this workspace?')) {
      deleteMutation.mutate(id)
    }
  }

  const workspaces = workspacesData?.workspaces || []

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>Workspaces</Title>
        <Button
          leftSection={<IconPlus size={16} />}
          onClick={() => setIsCreateModalOpen(true)}
        >
          Create Workspace
        </Button>
      </Group>

      {isLoading ? (
        <Text>Loading workspaces...</Text>
      ) : workspaces.length === 0 ? (
        <Card shadow="sm" padding="lg" radius="md" withBorder>
          <Stack align="center" py="xl">
            <IconUsers size={48} color="gray" />
            <Text size="lg" c="dimmed">No workspaces found</Text>
            <Text size="sm" c="dimmed">
              Create your first workspace to get started
            </Text>
          </Stack>
        </Card>
      ) : (
        <Grid>
          {workspaces.map((workspace) => (
            <Grid.Col key={workspace.id} span={{ base: 12, sm: 6, md: 4 }}>
              <Card shadow="sm" padding="lg" radius="md" withBorder>
                <Stack>
                  <Group justify="space-between">
                    <Text fw={500} size="lg">
                      {workspace.name}
                    </Text>
                    <Group>
                      <ActionIcon variant="subtle" color="blue">
                        <IconEdit size={16} />
                      </ActionIcon>
                      <ActionIcon
                        variant="subtle"
                        color="red"
                        onClick={() => handleDeleteWorkspace(workspace.id)}
                      >
                        <IconTrash size={16} />
                      </ActionIcon>
                    </Group>
                  </Group>
                  
                  <Text size="sm" c="dimmed">
                    Created: {new Date(workspace.created_at).toLocaleDateString()}
                  </Text>
                  
                  <Group>
                    <Badge variant="light" color="blue">
                      Owner
                    </Badge>
                  </Group>
                </Stack>
              </Card>
            </Grid.Col>
          ))}
        </Grid>
      )}

      <Modal
        opened={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Create New Workspace"
      >
        <Stack>
          <TextInput
            label="Workspace Name"
            placeholder="Enter workspace name"
            value={newWorkspaceName}
            onChange={(e) => setNewWorkspaceName(e.target.value)}
            required
          />
          <Group justify="flex-end">
            <Button variant="subtle" onClick={() => setIsCreateModalOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleCreateWorkspace}
              loading={createMutation.isPending}
            >
              Create
            </Button>
          </Group>
        </Stack>
      </Modal>
    </Stack>
  )
}

