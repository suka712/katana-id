import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldSeparator,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import logo from "/logo.svg";
import { useNavigate } from "react-router-dom";
import { useState } from "react";
import { useAuthStore } from "../store/useAuthStore";
import { LucideLoader2 } from "lucide-react";
import { useFormDebounce } from "@/hooks/use-form-debounce";
import { OAuthButtons } from "@/components/OAuthButtons";

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"form">) {
  const { login, isLoggingIn } = useAuthStore();
  const navigate = useNavigate();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  const { isDebouncing, shouldSubmit } = useFormDebounce({ email: "", password: "" });

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!shouldSubmit({ email, password })) return;

    await login({ email, password });

    if (useAuthStore.getState().token) {
      navigate("/dashboard");
    }
  };

  return (
    <form
      className={cn("flex flex-col gap-6 max-w-sm w-full", className)}
      {...props}
      onSubmit={handleSubmit}
    >
      <FieldGroup>
        <div className="flex flex-col items-center gap-1 text-center">
          <img src={logo} className="w-20"></img>
          <h1 className="text-2xl font-bold pt-5">Login to KatanaID</h1>
        </div>
        <Field>
          <FieldLabel htmlFor="email">Email</FieldLabel>
          <Input
            id="email"
            type="email"
            placeholder="damian@email.com"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </Field>
        <Field>
          <div className="flex items-center">
            <FieldLabel htmlFor="password">Password</FieldLabel>
            <a
              href="#"
              className="ml-auto text-sm underline-offset-4 hover:underline"
            >
              Forgot your password?
            </a>
          </div>
          <Input
            id="password"
            type="password"
            placeholder="••••••••"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </Field>
        <Field>
          <Button type="submit" disabled={isLoggingIn || isDebouncing}>
            {isLoggingIn ? <LucideLoader2 className="animate-spin" /> : "Login"}
          </Button>
        </Field>
        <FieldSeparator>Or continue with</FieldSeparator>
        <Field className="gap-5">
          <OAuthButtons labelPrefix="Login" />
          <FieldDescription className="text-center">
            Don&apos;t have an account?{" "}
            <a
              onClick={() => navigate("/signup")}
              className="underline underline-offset-4"
            >
              Sign up
            </a>
          </FieldDescription>
        </Field>
      </FieldGroup>
    </form>
  );
}
