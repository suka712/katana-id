import { LoginForm } from "@/components/login-form"
import GridBackground from "@/components/GridBackground"

const LoginPage = () => {
  return (
    <>
      <GridBackground
        glowColor="#a855f7"
        glowRadius={180}
        glowIntensity={0.3}
        gridSize={32}
      />
      <div className="relative z-10 flex justify-center items-center pt-10 md:pt-25 mx-10 md:mx-0">
        <LoginForm />
      </div>
    </>
  )
}

export default LoginPage