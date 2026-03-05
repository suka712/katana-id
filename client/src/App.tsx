import { Route, Routes } from "react-router-dom";
import { Toaster } from "sonner";
import LandingPage from "./pages/public-pages/LandingPage";
import LoginPage from "./pages/public-pages/LoginPage";
import SignupPage from "./pages/public-pages/SignupPage";

function App() {
  return (
    <>
      <Toaster position="top-center" />
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignupPage />} />
      </Routes>
    </>
  );
}

export default App;
