import { motion } from "framer-motion";
import { ReactNode } from "react";

interface AuthTransitionProps {
  children: ReactNode;
}

const AuthTransition = ({ children }: AuthTransitionProps) => {
  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: 20 }}
      transition={{ 
        duration: 0.3,
        ease: [0.22, 1, 0.36, 1] // Custom bezier curve for smooth transition
      }}
      className="w-full h-full"
    >
      {children}
    </motion.div>
  );
};

export default AuthTransition; 