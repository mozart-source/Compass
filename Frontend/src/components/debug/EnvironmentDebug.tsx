import React from "react";

interface EnvironmentInfo {
  environment: string;
  apiUrl: string;
  notesUrl: string;
  isDocker: boolean;
  isDevelopment: boolean;
  isProduction: boolean;
  nodeEnv: string;
}

const EnvironmentDebug: React.FC = () => {
  const [envInfo, setEnvInfo] = React.useState<EnvironmentInfo | null>(null);
  const [isVisible, setIsVisible] = React.useState(false);

  React.useEffect(() => {
    const getEnvironmentInfo = (): EnvironmentInfo => {
      const nodeEnv = import.meta.env.NODE_ENV || "development";
      const dockerEnv = import.meta.env.DOCKER_ENV === "true";

      return {
        environment: dockerEnv ? "Docker" : nodeEnv,
        apiUrl: import.meta.env.VITE_API_URL || "http://localhost:8081",
        notesUrl: import.meta.env.VITE_NOTES_URL || "http://localhost:5000",
        isDocker: dockerEnv,
        isDevelopment: nodeEnv === "development",
        isProduction: nodeEnv === "production",
        nodeEnv,
      };
    };

    setEnvInfo(getEnvironmentInfo());
  }, []);

  if (!envInfo) return null;

  return (
    <>
      {/* Toggle Button */}
      <button
        onClick={() => setIsVisible(!isVisible)}
        className="fixed bottom-4 right-4 z-50 bg-blue-600 hover:bg-blue-700 text-white p-2 rounded-full shadow-lg transition-colors"
        title="Toggle Environment Debug"
      >
        🔧
      </button>

      {/* Debug Panel */}
      {isVisible && (
        <div className="fixed bottom-16 right-4 z-50 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg shadow-xl p-4 max-w-sm">
          <div className="flex justify-between items-center mb-3">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Environment Debug
            </h3>
            <button
              onClick={() => setIsVisible(false)}
              className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            >
              ✕
            </button>
          </div>

          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="font-medium text-gray-700 dark:text-gray-300">
                Environment:
              </span>
              <span
                className={`px-2 py-1 rounded text-xs font-medium ${
                  envInfo.isDocker
                    ? "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
                    : envInfo.isDevelopment
                    ? "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200"
                    : "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                }`}
              >
                {envInfo.environment}
              </span>
            </div>

            <div className="flex justify-between">
              <span className="font-medium text-gray-700 dark:text-gray-300">
                Node ENV:
              </span>
              <span className="text-gray-600 dark:text-gray-400">
                {envInfo.nodeEnv}
              </span>
            </div>

            <div className="flex justify-between">
              <span className="font-medium text-gray-700 dark:text-gray-300">
                Docker:
              </span>
              <span
                className={envInfo.isDocker ? "text-blue-600" : "text-gray-600"}
              >
                {envInfo.isDocker ? "✓" : "✗"}
              </span>
            </div>

            <div className="pt-2 border-t border-gray-200 dark:border-gray-600">
              <div className="space-y-1">
                <div>
                  <span className="font-medium text-gray-700 dark:text-gray-300">
                    API URL:
                  </span>
                  <div className="text-xs text-gray-600 dark:text-gray-400 break-all">
                    {envInfo.apiUrl}
                  </div>
                </div>
                <div>
                  <span className="font-medium text-gray-700 dark:text-gray-300">
                    Notes URL:
                  </span>
                  <div className="text-xs text-gray-600 dark:text-gray-400 break-all">
                    {envInfo.notesUrl}
                  </div>
                </div>
              </div>
            </div>

            <div className="pt-2 border-t border-gray-200 dark:border-gray-600">
              <div className="text-xs text-gray-500 dark:text-gray-400">
                <div>Window: {typeof window !== "undefined" ? "✓" : "✗"}</div>
                <div>
                  Location:{" "}
                  {typeof window !== "undefined"
                    ? window.location.origin
                    : "N/A"}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default EnvironmentDebug;
