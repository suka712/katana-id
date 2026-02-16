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
import { useNavigate } from "react-router-dom";
import logo from "/logo.svg";
import { useState } from "react";
import { useAuthStore } from "../../store/useAuthStore";
import { toast } from "sonner";
import { LucideLoader2 } from "lucide-react";
import { useFormDebounce } from "@/hooks/use-form-debounce";
import { OAuthButtons } from "@/components/OAuthButtons";

export function SignupForm({
  className,
  ...props
}: React.ComponentProps<"form">) {
  const { signup, isSigningUp } = useAuthStore();
  const navigate = useNavigate();

  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isMatchedPassword, setIsMatchedPassword] = useState(true);

  const { isDebouncing, shouldSubmit } = useFormDebounce({
    username: "",
    email: "",
    password: "",
  });

  const isValidInfo = () => {
    if (username.trim().length < 3) {
      toast.error("Username must be at least 3 characters");
      return false;
    }

    if (password.trim().length < 8) {
      toast.error("Passwords must be at least 8 characters");
      return false;
    }

    setIsMatchedPassword(password === confirmPassword);
    if (password !== confirmPassword) {
      return false;
    }

    return true;
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!shouldSubmit({ username, email, password })) return;

    if (isValidInfo()) {
      await signup({ username, email, password });

      if (useAuthStore.getState().token) {
        navigate("/dashboard");
      }
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
          <h1 className="text-2xl font-bold pt-5">Create your Account</h1>
        </div>
        <Field>
          <FieldLabel htmlFor="name">Username</FieldLabel>
          <Input
            id="name"
            type="text"
            placeholder="Damian Reyes"
            required
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
        </Field>
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
          <FieldLabel htmlFor="password">Password</FieldLabel>
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
          <FieldLabel htmlFor="confirm-password">
            Confirm Password{" "}
            {!isMatchedPassword && (
              <span className="text-destructive font-normal">
                {" "}
                - Passwords do not match
              </span>
            )}
          </FieldLabel>
          <Input
            id="confirm-password"
            type="password"
            placeholder="••••••••"
            required
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
          />
        </Field>
        <Field>
          <Button type="submit" disabled={isSigningUp || isDebouncing}>
            {isSigningUp ? (
              <LucideLoader2 className="animate-spin" />
            ) : (
              "Create Account"
            )}
          </Button>
        </Field>
        <FieldSeparator>Or continue with</FieldSeparator>
        <Field className="gap-5">
          <OAuthButtons labelPrefix="Sign up" />
          <FieldDescription className="px-6 text-center">
            Already have an account?{" "}
            <a onClick={() => navigate("/login")}>Log in</a>
          </FieldDescription>
        </Field>
      </FieldGroup>
    </form>
  );
}
