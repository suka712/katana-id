import { Route, Routes } from "react-router-dom";
import { Toaster } from "sonner";
import LandingPage from "./pages/public-pages/LandingPage";
import LoginPage from "./pages/public-pages/LoginPage";
import SignupPage from "./pages/public-pages/SignupPage";
import GenerativeIdentityPage from "./pages/service-pages/GenerativeIdentityPage";
import TokenCallbackPage from "./pages/public-pages/AuthCallbackPage";
import DashboardLayout from "./components/layouts/DashboardLayout";
import RequireVerified from "./components/RequireVerified";

// const PublicLayout = () => (
//   <>
//     {/* Grid Background with cursor effect */}
//     <GridBackground
//       glowColor="#a855f7"
//       glowRadius={180}
//       glowIntensity={0.3}
//       gridSize={32}
//     />
//     <NavBar />
//     <Outlet />
//   </>
// );

function App() {
  return (
    <>
      <Toaster position="top-center" />
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignupPage />} />
        <Route path="/auth/callback" element={<TokenCallbackPage />} />
        <Route path="/auth/verified" element={<TokenCallbackPage />} />
        <Route path="/dashboard" element={<DashboardLayout />}>
          <Route
            index
            element={
              <RequireVerified>
                <GenerativeIdentityPage />
              </RequireVerified>
            }
          />
        </Route>
      </Routes>
    </>
  );
}

export default App;
