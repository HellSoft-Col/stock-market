#!/usr/bin/env python3
"""
Trading Simulation Runner
========================

Simple script to run the trading simulation with proper dependency checking.
This script will automatically install required dependencies if needed.

Usage:
    python3 run_simulation.py --tokens TK-1001,TK-1002,TK-1003 --duration 15
    python3 run_simulation.py --config simulation_config.json
"""

import sys
import subprocess
import json
import argparse
import os

def check_and_install_dependencies():
    """Check for required dependencies and install if missing"""
    required_packages = ['websockets']
    missing_packages = []
    
    for package in required_packages:
        try:
            __import__(package)
        except ImportError:
            missing_packages.append(package)
    
    if missing_packages:
        print(f"Missing packages: {', '.join(missing_packages)}")
        print("Installing required packages...")
        
        try:
            subprocess.check_call([
                sys.executable, '-m', 'pip', 'install', 
                *missing_packages
            ])
            print("Dependencies installed successfully!")
        except subprocess.CalledProcessError as e:
            print(f"Failed to install dependencies: {e}")
            print("Please install manually: pip install websockets")
            return False
    
    return True

def load_config(config_path):
    """Load configuration from JSON file"""
    try:
        with open(config_path, 'r') as f:
            return json.load(f)
    except FileNotFoundError:
        print(f"Configuration file not found: {config_path}")
        return None
    except json.JSONDecodeError as e:
        print(f"Invalid JSON in configuration file: {e}")
        return None

def run_simulation_script(tokens, server_url, duration, verbose=False):
    """Run the main trading simulation script"""
    cmd = [
        sys.executable, 
        'trading_simulation.py',
        '--tokens', ','.join(tokens),
        '--server', server_url,
        '--duration', str(duration)
    ]
    
    if verbose:
        cmd.append('--verbose')
    
    try:
        result = subprocess.run(cmd, check=True, cwd=os.path.dirname(__file__))
        return result.returncode == 0
    except subprocess.CalledProcessError as e:
        print(f"Simulation failed with exit code: {e.returncode}")
        return False

def main():
    parser = argparse.ArgumentParser(description='Trading Simulation Runner')
    parser.add_argument('--config', type=str, help='Configuration file path')
    parser.add_argument('--tokens', type=str, help='Comma-separated team tokens')
    parser.add_argument('--server', type=str, default='ws://localhost:8080', help='Server URL')
    parser.add_argument('--duration', type=int, default=15, help='Duration in minutes')
    parser.add_argument('--verbose', '-v', action='store_true', help='Verbose output')
    parser.add_argument('--install-deps', action='store_true', help='Install dependencies only')
    
    args = parser.parse_args()
    
    # Install dependencies if requested
    if args.install_deps:
        return 0 if check_and_install_dependencies() else 1
    
    # Check dependencies
    if not check_and_install_dependencies():
        return 1
    
    # Load configuration
    if args.config:
        config = load_config(args.config)
        if not config:
            return 1
        
        tokens = config.get('tokens', [])
        server_url = config.get('simulation', {}).get('server_url', 'ws://localhost:8080')
        duration = config.get('simulation', {}).get('duration_minutes', 15)
        verbose = config.get('simulation', {}).get('log_level', 'INFO') == 'DEBUG'
    else:
        if not args.tokens:
            print("Error: --tokens or --config is required")
            return 1
        
        tokens = [token.strip() for token in args.tokens.split(',')]
        server_url = args.server
        duration = args.duration
        verbose = args.verbose
    
    if len(tokens) < 2:
        print("Error: At least 2 tokens are required for simulation")
        return 1
    
    print(f"Starting simulation with {len(tokens)} clients for {duration} minutes")
    print(f"Server: {server_url}")
    print(f"Tokens: {', '.join(tokens)}")
    
    # Run simulation
    success = run_simulation_script(tokens, server_url, duration, verbose)
    return 0 if success else 1

if __name__ == "__main__":
    exit_code = main()
    sys.exit(exit_code)