import React, { useState } from 'react'
import { AppShell, Navbar, Header, Text, Button, Group, Container } from '@mantine/core'
import { IconCloud, IconSettings, IconUsers, IconServer } from '@tabler/icons-react'
import { useAuth } from './hooks/useAuth'
import LoginForm from './components/LoginForm'
import WorkspaceList from './components/WorkspaceList'
import VMList from './components/VMList'
import CredentialsList from './components/CredentialsList'
import TofuExecutions from './components/TofuExecutions'

function App() {
  const { user, login, logout, isAuthenticated } = useAuth()
  const [activeTab, setActiveTab] = useState('workspaces')

  if (!isAuthenticated) {
    return (
      <Container size="sm" style={{ marginTop: '10%' }}>
        <LoginForm onLogin={login} />
      </Container>
    )
  }

  const renderContent = () => {
    switch (activeTab) {
      case 'workspaces':
        return <WorkspaceList />
      case 'vms':
        return <VMList />
      case 'credentials':
        return <CredentialsList />
      case 'tofu':
        return <TofuExecutions />
      default:
        return <WorkspaceList />
    }
  }

  return (
    <AppShell
      navbar={{
        width: 250,
        breakpoint: 'sm',
        children: (
          <Navbar p="md">
            <Navbar.Section>
              <Group mb="xl">
                <IconCloud size={24} />
                <Text size="lg" fw={700}>CMP</Text>
              </Group>
            </Navbar.Section>
            
            <Navbar.Section grow>
              <Button
                variant={activeTab === 'workspaces' ? 'filled' : 'subtle'}
                leftSection={<IconUsers size={16} />}
                fullWidth
                justify="flex-start"
                mb="xs"
                onClick={() => setActiveTab('workspaces')}
              >
                Workspaces
              </Button>
              
              <Button
                variant={activeTab === 'vms' ? 'filled' : 'subtle'}
                leftSection={<IconServer size={16} />}
                fullWidth
                justify="flex-start"
                mb="xs"
                onClick={() => setActiveTab('vms')}
              >
                Virtual Machines
              </Button>
              
              <Button
                variant={activeTab === 'credentials' ? 'filled' : 'subtle'}
                leftSection={<IconSettings size={16} />}
                fullWidth
                justify="flex-start"
                mb="xs"
                onClick={() => setActiveTab('credentials')}
              >
                Credentials
              </Button>
              
              <Button
                variant={activeTab === 'tofu' ? 'filled' : 'subtle'}
                leftSection={<IconSettings size={16} />}
                fullWidth
                justify="flex-start"
                mb="xs"
                onClick={() => setActiveTab('tofu')}
              >
                OpenTofu
              </Button>
            </Navbar.Section>
            
            <Navbar.Section>
              <Text size="sm" c="dimmed" mb="xs">
                {user?.email}
              </Text>
              <Button variant="subtle" fullWidth onClick={logout}>
                Logout
              </Button>
            </Navbar.Section>
          </Navbar>
        ),
      }}
      header={{
        height: 60,
        children: (
          <Header p="md">
            <Group justify="space-between">
              <Text size="xl" fw={700}>
                Cloud Management Platform
              </Text>
              <Text size="sm" c="dimmed">
                Welcome, {user?.name}
              </Text>
            </Group>
          </Header>
        ),
      }}
    >
      <Container size="xl" p="md">
        {renderContent()}
      </Container>
    </AppShell>
  )
}

export default App

