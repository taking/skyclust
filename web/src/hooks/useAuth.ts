import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notifications } from '@mantine/notifications'
import { authApi, User, LoginRequest, RegisterRequest } from '../api/auth'

export const useAuth = () => {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const queryClient = useQueryClient()

  // Check if user is authenticated on mount
  useEffect(() => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      setIsAuthenticated(true)
    }
  }, [])

  // Get current user query
  const { data: userData, isLoading } = useQuery({
    queryKey: ['currentUser'],
    queryFn: authApi.getCurrentUser,
    enabled: isAuthenticated,
    retry: false,
  })

  // Login mutation
  const loginMutation = useMutation({
    mutationFn: authApi.login,
    onSuccess: (data) => {
      localStorage.setItem('auth_token', data.token)
      setIsAuthenticated(true)
      queryClient.setQueryData(['currentUser'], { user: data.user })
      notifications.show({
        title: 'Success',
        message: 'Logged in successfully',
        color: 'green',
      })
    },
    onError: (error: any) => {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.error || 'Login failed',
        color: 'red',
      })
    },
  })

  // Register mutation
  const registerMutation = useMutation({
    mutationFn: authApi.register,
    onSuccess: () => {
      notifications.show({
        title: 'Success',
        message: 'Account created successfully',
        color: 'green',
      })
    },
    onError: (error: any) => {
      notifications.show({
        title: 'Error',
        message: error.response?.data?.error || 'Registration failed',
        color: 'red',
      })
    },
  })

  // Login function
  const login = useCallback((data: LoginRequest) => {
    loginMutation.mutate(data)
  }, [loginMutation])

  // Register function
  const register = useCallback((data: RegisterRequest) => {
    registerMutation.mutate(data)
  }, [registerMutation])

  // Logout function
  const logout = useCallback(() => {
    localStorage.removeItem('auth_token')
    setIsAuthenticated(false)
    queryClient.clear()
    notifications.show({
      title: 'Success',
      message: 'Logged out successfully',
      color: 'blue',
    })
  }, [queryClient])

  return {
    user: userData?.user,
    isAuthenticated,
    isLoading: isLoading || loginMutation.isPending || registerMutation.isPending,
    login,
    register,
    logout,
  }
}

