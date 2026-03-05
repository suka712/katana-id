import { create } from "zustand";
import { persist } from "zustand/middleware";
import { axiosInstance } from "../lib/axios";
import { AxiosError } from "axios";
import { toast } from "sonner";

interface AuthUser {
  username: string;
  email: string;
}

interface AuthStore {
  authUser: AuthUser | null;
  token: string | null;
  isSigningUp: boolean;
  isLoggingIn: boolean;
  signup: (data: { username: string; email: string; password: string }) => Promise<void>;
  login: (data: { email: string; password: string }) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      authUser: null,
      token: null,
      isSigningUp: false,
      isLoggingIn: false,

      signup: async (data) => {
        set({ isSigningUp: true });
        try {
          const res = await axiosInstance.post("/auth/signup", data);
          set({ token: res.data.token, authUser: { username: res.data.username, email: res.data.email } });
          toast.success("Account created successfully.");
        } catch (error: unknown) {
          if (error instanceof AxiosError) {
            toast.error(error.response?.status === 429
              ? "Too many attempts. Try again later."
              : "Error signing up: " + error.response?.data.error);
          } else {
            toast.error("Error creating account.");
          }
        } finally {
          set({ isSigningUp: false });
        }
      },

      login: async (data) => {
        set({ isLoggingIn: true });
        try {
          const res = await axiosInstance.post("/auth/login", data);
          set({ token: res.data.token, authUser: { username: res.data.username, email: res.data.email } });
          toast.success("Logged in successfully.");
        } catch (error: unknown) {
          if (error instanceof AxiosError) {
            toast.error(error.response?.status === 429
              ? "Too many attempts. Try again later."
              : "Error logging in: " + error.response?.data.error);
          } else {
            toast.error("Error logging in.");
          }
        } finally {
          set({ isLoggingIn: false });
        }
      },

      logout: () => {
        set({ token: null, authUser: null });
        toast.success("Successfully logged out");
      },
    }),
    { name: "auth-storage" }
  )
);
