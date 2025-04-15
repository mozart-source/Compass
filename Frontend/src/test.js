const WebSocket = require('ws');

const token = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNDFlMTVjMzEtNWRkYS00M2Q4LWFlYTUtMmQ2MDQ5OTM2YzFkIiwiZW1haWwiOiJhQGEuY29tIiwicm9sZXMiOlsidXNlciJdLCJvcmdfaWQiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAiLCJwZXJtaXNzaW9ucyI6WyJ0YXNrczpyZWFkIiwidGFza3M6dXBkYXRlIiwidGFza3M6Y3JlYXRlIiwicHJvamVjdHM6cmVhZCIsIm9yZ2FuaXphdGlvbnM6cmVhZCJdLCJleHAiOjE3NDg0MzA1MDYsIm5iZiI6MTc0ODM0NDEwNiwiaWF0IjoxNzQ4MzQ0MTA2fQ.uy_l5bP44JtNKx9eWIE4xNpbxYmoRMX25ghUyL-PYx8';

// Pass token as a query parameter instead of a header
const ws = new WebSocket(`ws://localhost:8000/api/notifications/ws?token=${token}`);

console.log('Listening for notifications. Press Ctrl+C to exit.');

// Store received notifications
const notifications = [];

ws.on('open', function open() {
  console.log('WebSocket connection established');
  
  // Set up command input from console
  process.stdin.on('data', (data) => {
    const input = data.toString().trim();
    
    if (input === 'exit' || input === 'quit') {
      console.log('Closing connection...');
      ws.close();
      process.exit(0);
    } else if (input === 'mark-all-read') {
      // Send command to mark all notifications as read
      ws.send(JSON.stringify({
        command: 'mark_all_read'
      }));
      console.log('Sent command: mark all as read');
    } else if (input.startsWith('mark-read ')) {
      // Extract notification ID and send mark-read command
      const id = input.substring('mark-read '.length).trim();
      ws.send(JSON.stringify({
        command: 'mark_read',
        id: id
      }));
      console.log(`Sent command: mark notification ${id} as read`);
    } else if (input === 'list') {
      // List all received notifications
      console.log('Received notifications:');
      notifications.forEach((notif, index) => {
        console.log(`[${index}] ID: ${notif.id} - ${notif.title}: ${notif.content}`);
      });
    } else if (input === 'help') {
      console.log('\nAvailable commands:');
      console.log('  help - Show this help message');
      console.log('  list - List all received notifications');
      console.log('  mark-read <id> - Mark specific notification as read');
      console.log('  mark-all-read - Mark all notifications as read');
      console.log('  exit/quit - Close the connection and exit\n');
    } else {
      console.log('Unknown command. Type "help" for available commands.');
    }
  });
  
  console.log('\nType "help" for available commands\n');
});

ws.on('message', function incoming(data) {
  try {
    const notification = JSON.parse(data.toString());
    
    // Check if it's a notification or a count message
    if (notification.type === 'count') {
      console.log(`You have ${notification.count} unread notifications`);
    } else if (notification.id) {
      // It's a notification
      console.log('Notification received:');
      console.log(`ID: ${notification.id}`);
      console.log(`Title: ${notification.title}`);
      console.log(`Content: ${notification.content}`);
      console.log(`Type: ${notification.type}`);
      console.log(`Status: ${notification.status}`);
      console.log('-----------------------------------');
      
      // Store the notification
      notifications.push(notification);
    }
  } catch (err) {
    console.error('Error parsing notification:', data.toString());
  }
});

ws.on('close', function close() {
  console.log('WebSocket connection closed');
  process.exit(0);
});

ws.on('error', function error(err) {
  console.error('WebSocket error:', err);
  process.exit(1);
});