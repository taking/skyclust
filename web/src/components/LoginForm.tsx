import React, { useState } from 'react'
import {
  Paper,
  TextInput,
  PasswordInput,
  Button,
  Stack,
  Title,
  Text,
  Tabs,
  Group,
  Divider,
} from '@mantine/core'
import { IconMail, IconLock, IconUser } from '@tabler/icons-react'
import { LoginRequest, RegisterRequest } from '../api/auth'

interface LoginFormProps {
  onLogin: (data: LoginRequest) => void
  onRegister?: (data: RegisterRequest) => void
}

export default function LoginForm({ onLogin, onRegister }: LoginFormProps) {
  const [activeTab, setActiveTab] = useState<string | null>('login')
  const [loginData, setLoginData] = useState<LoginRequest>({
    email: '',
    password: '',
  })
  const [registerData, setRegisterData] = useState<RegisterRequest>({
    email: '',
    password: '',
    name: '',
  })

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault()
    onLogin(loginData)
  }

  const handleRegister = (e: React.FormEvent) => {
    e.preventDefault()
    if (onRegister) {
      onRegister(registerData)
    }
  }

  return (
    <Paper shadow="md" p="xl" radius="md" style={{ width: '100%', maxWidth: 400 }}>
      <Stack align="center" mb="xl">
        <Title order={2}>Welcome to CMP</Title>
        <Text c="dimmed" size="sm">
          Cloud Management Platform
        </Text>
      </Stack>

      <Tabs value={activeTab} onTabChange={setActiveTab}>
        <Tabs.List>
          <Tabs.Tab value="login">Login</Tabs.Tab>
          <Tabs.Tab value="register">Register</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="login" pt="md">
          <form onSubmit={handleLogin}>
            <Stack>
              <TextInput
                label="Email"
                placeholder="your@email.com"
                leftSection={<IconMail size={16} />}
                value={loginData.email}
                onChange={(e) => setLoginData({ ...loginData, email: e.target.value })}
                required
              />
              <PasswordInput
                label="Password"
                placeholder="Your password"
                leftSection={<IconLock size={16} />}
                value={loginData.password}
                onChange={(e) => setLoginData({ ...loginData, password: e.target.value })}
                required
              />
              <Button type="submit" fullWidth>
                Login
              </Button>
            </Stack>
          </form>
        </Tabs.Panel>

        <Tabs.Panel value="register" pt="md">
          <form onSubmit={handleRegister}>
            <Stack>
              <TextInput
                label="Name"
                placeholder="Your name"
                leftSection={<IconUser size={16} />}
                value={registerData.name}
                onChange={(e) => setRegisterData({ ...registerData, name: e.target.value })}
                required
              />
              <TextInput
                label="Email"
                placeholder="your@email.com"
                leftSection={<IconMail size={16} />}
                value={registerData.email}
                onChange={(e) => setRegisterData({ ...registerData, email: e.target.value })}
                required
              />
              <PasswordInput
                label="Password"
                placeholder="Your password"
                leftSection={<IconLock size={16} />}
                value={registerData.password}
                onChange={(e) => setRegisterData({ ...registerData, password: e.target.value })}
                required
              />
              <Button type="submit" fullWidth>
                Register
              </Button>
            </Stack>
          </form>
        </Tabs.Panel>
      </Tabs>
    </Paper>
  )
}

