import React from 'react';

const Home: React.FC = () => {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold text-gray-900 mb-4">Welcome to Your Dashboard</h1>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div className="bg-white p-4 rounded-lg shadow">
          <h2 className="text-lg font-semibold mb-2">Quick Overview</h2>
          <p className="text-gray-600">View your daily summary and upcoming tasks</p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h2 className="text-lg font-semibold mb-2">Recent Activities</h2>
          <p className="text-gray-600">Check your latest calendar events and todos</p>
        </div>
        <div className="bg-white p-4 rounded-lg shadow">
          <h2 className="text-lg font-semibold mb-2">Quick Actions</h2>
          <p className="text-gray-600">Create new tasks or schedule meetings</p>
        </div>
      </div>
    </div>
  );
};

export default Home; 