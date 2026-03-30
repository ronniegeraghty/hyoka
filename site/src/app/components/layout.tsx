import { Outlet } from "react-router";
import { Navbar } from "./navbar";
import { Footer } from "./footer";

export function Layout() {
  return (
    <div className="min-h-screen bg-[#0a0a0f]">
      <Navbar />
      <Outlet />
      <Footer />
    </div>
  );
}
