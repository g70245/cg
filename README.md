# Multi-Game Instance Bot

A versatile automation tool developed using **Golang**, **Fyne**, and **Win32 API** to manage multiple game instances with advanced features and human-like behavior simulation.


## Features  

### 1. Programmable Combat Actions  
Enjoy the flexibility of programmable combat actions, allowing you to customize and optimize your in-game strategies.  

### 2. Randomized Movement within a Specified Range  
This bot incorporates a smart movement algorithm, randomly navigating within predefined ranges. This feature is designed to mimic human-like behavior, making it less susceptible to detection as a bot.  

### 3. Audible Cues for Seamless Control  
Improve your control over the game with audible cues. Receive prompt audio notifications to signal when it's time to return to the game and take manual control.  

### 4. Semi-Automatic Item Production and Stacking  
Optimize your gameplay with semi-automatic item production. The bot automatically generates items and stacks them until your inventory is full, ensuring efficient resource management.  

#### Sound Notifications  
- **Injured Alert:** Receive a sound notification when your character is injured, prompting you to take necessary actions.  
- **Inventory Full Warning:** Be alerted with a sound notification when your inventory is full.  

### 5. Unlimited Simultaneous Control  
Take advantage of your computer's capabilities to the fullest. There's no limit to the number of games you can control simultaneously, allowing you to fully utilize your system resources.  


## Getting Started  

### Prerequisites  
Before getting started, make sure you have the following installed:  
1. **Fyne:** Install Fyne by following the instructions [here](https://developer.fyne.io/started/).  
2. **Golang:** Install Golang by following the instructions [here](https://go.dev/doc/install).  

### Running and Building  
```bash
go mod download

# To run the application directly, use the following command:
go run .

# If you prefer to build the executable first and then run it, use the following commands:
go build .

# If you want to package the application for Windows using the fyne tool, use the following command:
fyne package -os windows -icon C:\path\to\your\icon.png
```

## Example
![alt text](https://github.com/g70245/cg/blob/main/example.png?raw=true)  
*Example interface showing the bot in action, demonstrating customizable settings and monitoring features.*
