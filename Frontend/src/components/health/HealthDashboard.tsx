import React, { useState } from 'react';

interface HealthMetrics {
  steps: number;
  heartRate: number;
  calories: number;
  sleep: number;
}

const HealthDashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<HealthMetrics>({
    steps: 8432,
    heartRate: 72,
    calories: 1850,
    sleep: 7.5,
  });

  return (
    <div className="min-h-screen bg-gray-100 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Health Monitor</h1>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {/* Steps Card */}
          <div className="bg-white rounded-xl shadow-md p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-700">Steps</h2>
              <span className="text-blue-500">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
                </svg>
              </span>
            </div>
            <p className="text-3xl font-bold text-gray-900">{metrics.steps.toLocaleString()}</p>
            <p className="text-sm text-gray-500 mt-2">Daily Goal: 10,000</p>
          </div>

          {/* Heart Rate Card */}
          <div className="bg-white rounded-xl shadow-md p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-700">Heart Rate</h2>
              <span className="text-red-500">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                </svg>
              </span>
            </div>
            <p className="text-3xl font-bold text-gray-900">{metrics.heartRate} <span className="text-sm">bpm</span></p>
            <p className="text-sm text-gray-500 mt-2">Resting Heart Rate</p>
          </div>

          {/* Calories Card */}
          <div className="bg-white rounded-xl shadow-md p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-700">Calories</h2>
              <span className="text-orange-500">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                </svg>
              </span>
            </div>
            <p className="text-3xl font-bold text-gray-900">{metrics.calories}</p>
            <p className="text-sm text-gray-500 mt-2">Calories Burned</p>
          </div>

          {/* Sleep Card */}
          <div className="bg-white rounded-xl shadow-md p-6">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-gray-700">Sleep</h2>
              <span className="text-purple-500">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
                </svg>
              </span>
            </div>
            <p className="text-3xl font-bold text-gray-900">{metrics.sleep} <span className="text-sm">hrs</span></p>
            <p className="text-sm text-gray-500 mt-2">Last Night's Sleep</p>
          </div>
        </div>

        <div className="mt-8 bg-white rounded-xl shadow-md p-6">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Weekly Activity</h2>
          <div className="h-64 bg-gray-50 rounded-lg">
            {/* TODO: Add chart component here */}
            <div className="flex items-center justify-center h-full text-gray-400">
              Chart placeholder - Weekly activity visualization will be displayed here
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HealthDashboard;
