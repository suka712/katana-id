"use client";

import React from "react";
import { motion } from "framer-motion";
import { cn } from "../../lib/utils";

export const LampContainer = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return (
    <div
      className={cn(
        "relative flex min-h-[10vh] flex-col items-center justify-center w-full rounded-md z-0",
        className
      )}
    >
      <div className="relative flex w-full flex-1 scale-y-125 items-center justify-center z-0">
        {/* Wide violet-cyan gradient glow bar */}
        <motion.div
          initial={{ width: "12rem", opacity: 0.5 }}
          whileInView={{ width: "28rem", opacity: 1 }}
          transition={{
            delay: 0.3,
            duration: 0.9,
            ease: "easeInOut",
          }}
          className="absolute inset-auto z-30 h-6 bg-linear-to-r from-transparent via-violet-500 to-transparent blur-md"
        />
        {/* Thin crisp centerline */}
        <motion.div
          initial={{ width: "18rem" }}
          whileInView={{ width: "42rem" }}
          transition={{
            delay: 0.3,
            duration: 0.9,
            ease: "easeInOut",
          }}
          className="absolute inset-auto z-50 h-px bg-linear-to-r from-transparent via-violet-300 to-transparent shadow-[0_0_12px_3px_rgba(167,139,250,0.7)]"
        />
        {/* Ambient glow cone below */}
        <motion.div
          initial={{ opacity: 0, height: "0px" }}
          whileInView={{ opacity: 1, height: "180px" }}
          transition={{ delay: 0.5, duration: 0.8, ease: "easeOut" }}
          className="absolute inset-auto z-20 top-1/2 w-[28rem] bg-[conic-gradient(from_180deg_at_50%_0%,transparent_25%,oklch(65%_0.22_268/0.12)_50%,transparent_75%)] blur-2xl"
        />
      </div>
      <div className="relative z-50 flex flex-col items-center px-5">
        {children}
      </div>
    </div>
  );
};
