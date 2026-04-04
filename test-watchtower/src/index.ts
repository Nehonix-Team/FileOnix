
function greet(name: string) {
  console.log(`Hello, ${name}!`);
  console.log("The current time is:", new Date().toLocaleTimeString());
}
 
// Run the greeting every second
setInterval(() => {
  greet("Developer");
}, 1000); 

console.log("Server started! Edit this file to see the changes.");






