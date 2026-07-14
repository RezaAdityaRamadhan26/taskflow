export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
  created_at: string;
}

export interface Workspace {
  id: string;
  name: string;
  description?: string;
  slug: string;
  role: 'OWNER' | 'ADMIN' | 'MEMBER';
  created_at: string;
}

export interface WorkspaceMember {
  id: string;
  user_id: string;
  name: string;
  email: string;
  avatar_url?: string;
  role: 'OWNER' | 'ADMIN' | 'MEMBER';
  joined_at: string;
}

export interface Board {
  id: string;
  workspace_id: string;
  name: string;
  description?: string;
  color?: string;
  created_at: string;
}

export interface List {
  id: string;
  board_id: string;
  name: string;
  position: number;
}

export interface Card {
  id: string;
  list_id: string;
  title: string;
  description?: string;
  position: number;
  priority: 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT';
  due_date?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
}
