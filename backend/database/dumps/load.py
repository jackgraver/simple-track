#!/usr/bin/env python3
"""
Script to parse workout log text file and generate SQL INSERT statements.
"""

import re
from datetime import datetime
from typing import List, Tuple, Optional

# Exercise name mappings from text file to database
EXERCISE_MAP = {
    "incline press": 1,
    "fly": 2,
    "chest fly": 2,
    "dips": 3,
    "lat raise": 4,
    "lay raise": 4,  # typo variant
    "shoulder press": 5,
    "squat": 6,
    "squad": 6,  # typo variant
    "deadlift": 7,
    "deadlifts": 7,
    "bench press": 8,
    "flat bench": 8,
    "cable rows": 9,
    "low rows": 9,
    "barbell rows": 10,
    "face pulls": 11,
    "facepulls": 11,
    "fave pulls": 11,  # typo variant
    "pulldowns": 12,
    "jm press": 13,
    "jm": 13,
    "extensions": 14,
    "pushdown": 14,
    "pushdown (v)": 14,
    "incline curls": 15,
    "hammer curls": 16,
    "hammer": 16,
    "calf raises": 17,
    "calfs": 17,
    "standing calfs": 17,
    "ab crunches": 18,
    "abs crunch": 18,
    "leg press": 19,
    "leg extensions": 20,
    "hamstring curls": 21,
    "hamstring curl": 21,
    "hip press": 22,
    "hip extensions": 23,
    "hip hinge": 23,
    "outer thigh": 24,
    "inner thigh": 25,
    # Additional exercises that might not be in database
    "pull-ups": None,  # Not in database, will skip
    "pull ups": None,
    "rear delt": None,
    "t-bar/kelso": None,
    "t-bar": None,
    "kelso": None,
    "shrugs": None,
    "skullcrushers": None,
    "single arm push down": None,
    "back extensions": None,
    "hack squat": None,
    "forearms": None,
    "bayesian": None,
}

# Workout plan mappings
WORKOUT_PLAN_MAP = {
    "push": 1,
    "pull": 2,
    "legs": 3,
    "lower": 5,  # Lower is workout plan 5
    "upper": 4,
    "active rest": 6,
    "rest": 7,
}

# Month abbreviations
MONTHS = {
    "jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
    "jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12
}

def parse_date(date_str: str, year: int = 2025) -> Optional[datetime]:
    """Parse date string like 'Oct 13' or 'Nov 4' into datetime."""
    date_str = date_str.strip().lower()
    match = re.match(r"([a-z]+)\s+(\d+)", date_str)
    if not match:
        return None
    
    month_str, day_str = match.groups()
    month = MONTHS.get(month_str[:3])
    if not month:
        return None
    
    try:
        day = int(day_str)
        return datetime(year, month, day)
    except ValueError:
        return None

def parse_workout_type(workout_str: str) -> Optional[int]:
    """Extract workout type from string like '(push)' or '(Upper)'."""
    match = re.search(r"\(([^)]+)\)", workout_str)
    if not match:
        return None
    
    workout_type = match.group(1).strip().lower()
    return WORKOUT_PLAN_MAP.get(workout_type)

def normalize_exercise_name(name: str) -> Optional[int]:
    """Normalize exercise name and return exercise ID."""
    # Remove extra whitespace and convert to lowercase
    name = re.sub(r"\s+", " ", name.strip().lower())
    
    # Remove parenthetical notes like "(physio)", "(V)", etc.
    name = re.sub(r"\s*\([^)]+\)", "", name)
    
    # Remove trailing dashes and special characters
    name = name.rstrip(" -")
    
    # Handle specific name variations
    if "machine" in name:
        if "lat raise" in name:
            name = "lat raise"
        elif "raise" in name:
            name = "lat raise"
    
    if "dumbbell" in name and "lat raise" in name:
        name = "lat raise"
    
    if "single arm" in name and "push" in name:
        name = "extensions"  # Map to Extensions
    
    if "skullcrushers" in name or "skull crushers" in name:
        name = "extensions"  # Map to Extensions
    
    if "hack squat" in name:
        name = "squat"
    
    if "standing" in name and "calf" in name:
        name = "calfs"
    
    # Try exact match first
    if name in EXERCISE_MAP:
        return EXERCISE_MAP[name]
    
    # Try partial matches (check if key is contained in name or vice versa)
    for key, exercise_id in EXERCISE_MAP.items():
        if exercise_id is None:
            continue
        if key in name or name in key:
            return exercise_id
    
    return None

def parse_weight(weight_str: str) -> Optional[float]:
    """Parse weight string, handling formats like '25s', '-100', '40', etc."""
    weight_str = weight_str.strip().lower()
    
    # Remove 's' suffix (e.g., "25s" -> 25)
    if weight_str.endswith('s'):
        weight_str = weight_str[:-1]
    
    # Handle negative weights (assisted exercises)
    try:
        weight = float(weight_str)
        return weight
    except ValueError:
        return None

def parse_reps(reps_str: str) -> Optional[int]:
    """Parse reps string, handling formats like '7', '6?', etc."""
    reps_str = reps_str.strip()
    
    # Remove question marks and other trailing characters
    reps_str = re.sub(r"[?+\s]+$", "", reps_str)
    
    try:
        return int(reps_str)
    except ValueError:
        return None

def parse_sets(sets_str: str) -> List[Tuple[float, int]]:
    """Parse sets string like '40x7, 40x6' or '25sx8'."""
    sets = []
    last_reps = None
    
    # Remove trailing notes in parentheses like "(form reset)", "(off day)", etc.
    sets_str = re.sub(r"\s*\([^)]+\)\s*$", "", sets_str)
    
    # Split by comma
    set_parts = [s.strip() for s in sets_str.split(",")]
    
    for set_part in set_parts:
        # Skip empty parts
        if not set_part:
            continue
        
        # Remove trailing special characters like "+++", "?", etc.
        set_part = re.sub(r"[?+\s]+$", "", set_part)
        
        # Handle formats like "40x7", "25sx8", "-100x6", "40x" (no reps)
        # Also handle "25x6/7" (take first number)
        set_part = set_part.split("/")[0].strip()
        
        # Handle "->" notation (e.g., "70x11 ->80") - take the first part
        if "->" in set_part:
            set_part = set_part.split("->")[0].strip()
        
        match = re.match(r"([-\d.]+s?)\s*x\s*(\d*)", set_part, re.IGNORECASE)
        if not match:
            continue
        
        weight_str, reps_str = match.groups()
        
        weight = parse_weight(weight_str)
        if weight is None:
            continue
        
        # If no reps specified, use last reps or skip
        if not reps_str:
            if last_reps is None:
                continue
            reps = last_reps
        else:
            reps = parse_reps(reps_str)
            if reps is None:
                continue
            last_reps = reps
        
        sets.append((weight, reps))
    
    return sets

def parse_workout_file(filepath: str) -> List[dict]:
    """Parse the workout log text file and return structured data."""
    workouts = []
    current_workout = None
    current_year = 2025  # Default year, will adjust based on dates
    
    with open(filepath, 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    i = 0
    while i < len(lines):
        line = lines[i].strip()
        
        # Skip empty lines
        if not line:
            i += 1
            continue
        
        # Check if this is a date/workout type line
        date_match = re.match(r"([A-Za-z]+\s+\d+)\s*\(([^)]+)\)", line)
        if date_match:
            # Save previous workout if exists
            if current_workout and current_workout['exercises']:
                workouts.append(current_workout)
            
            date_str, workout_type_str = date_match.groups()
            
            # Try to determine year based on month
            # If we see Nov/Dec and current year is 2025, might need to check
            # For now, assume all dates are in 2025
            date = parse_date(date_str, current_year)
            workout_plan_id = parse_workout_type(f"({workout_type_str})")
            
            if date and workout_plan_id:
                current_workout = {
                    'date': date,
                    'workout_plan_id': workout_plan_id,
                    'exercises': []
                }
            else:
                current_workout = None
            i += 1
            continue
        
        # Check if this is an exercise line (contains " - " or ends with "x" followed by numbers)
        if current_workout and (" - " in line or re.search(r"\s*x\s*\d", line)):
            # Parse exercise name and sets
            if " - " in line:
                parts = line.split(" - ", 1)
                exercise_name = parts[0].strip()
                sets_str = parts[1].strip() if len(parts) > 1 else ""
            else:
                # Some lines might just have "Exercise -" with no sets
                if " -" in line:
                    exercise_name = line.split(" -")[0].strip()
                    sets_str = ""
                else:
                    # Try to find where exercise name ends and sets begin
                    match = re.match(r"(.+?)\s+([-\d.]+s?\s*x\s*\d)", line)
                    if match:
                        exercise_name = match.group(1).strip()
                        sets_str = match.group(2).strip()
                    else:
                        i += 1
                        continue
            
            exercise_id = normalize_exercise_name(exercise_name)
            
            if exercise_id is None:
                # Exercise not found in database, skip
                i += 1
                continue
            
            sets = parse_sets(sets_str) if sets_str.strip() else []
            
            # Only add exercise if it has sets
            if sets:
                current_workout['exercises'].append({
                    'exercise_id': exercise_id,
                    'sets': sets
                })
        
        i += 1
    
    # Don't forget the last workout
    if current_workout and current_workout['exercises']:
        workouts.append(current_workout)
    
    return workouts

def generate_sql(workouts: List[dict], start_workout_log_id: int = 1000, start_exercise_id: int = 1000, start_set_id: int = 1000) -> str:
    """Generate SQL INSERT statements for the parsed workouts."""
    sql_lines = []
    
    workout_log_id = start_workout_log_id
    logged_exercise_id = start_exercise_id
    logged_set_id = start_set_id
    
    # Get current timestamp for created_at/updated_at
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    
    for workout in workouts:
        # Insert workout log
        date_str = workout['date'].strftime("%Y-%m-%d %H:%M:%S")
        sql_lines.append(
            f"INSERT INTO workout_logs (id, created_at, updated_at, deleted_at, date, workout_plan_id) "
            f"VALUES({workout_log_id}, '{timestamp}', '{timestamp}', NULL, '{date_str}', {workout['workout_plan_id']});"
        )
        
        # Insert logged exercises and sets
        for exercise in workout['exercises']:
            sql_lines.append(
                f"INSERT INTO logged_exercises (id, created_at, updated_at, deleted_at, workout_log_id, exercise_id, weight_setup) "
                f"VALUES({logged_exercise_id}, '{timestamp}', '{timestamp}', NULL, {workout_log_id}, {exercise['exercise_id']}, '');"
            )
            
            for weight, reps in exercise['sets']:
                sql_lines.append(
                    f"INSERT INTO logged_sets (id, created_at, updated_at, deleted_at, logged_exercise_id, reps, weight) "
                    f"VALUES({logged_set_id}, '{timestamp}', '{timestamp}', NULL, {logged_exercise_id}, {reps}, {weight});"
                )
                logged_set_id += 1
            
            logged_exercise_id += 1
        
        workout_log_id += 1
    
    return "\n".join(sql_lines)

def main():
    """Main function to parse file and generate SQL."""
    input_file = "data_week.txt"
    output_file = "workout_logs_insert_week.sql"
    
    print(f"Parsing {input_file}...")
    workouts = parse_workout_file(input_file)
    
    print(f"Found {len(workouts)} workouts")
    total_exercises = sum(len(w['exercises']) for w in workouts)
    total_sets = sum(sum(len(e['sets']) for e in w['exercises']) for w in workouts)
    print(f"Total exercises: {total_exercises}, Total sets: {total_sets}")
    
    print(f"Generating SQL...")
    sql = generate_sql(workouts)
    
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write("-- Generated SQL INSERT statements for workout logs\n")
        f.write("-- Generated from data_week.txt\n\n")
        f.write(sql)
    
    print(f"SQL written to {output_file}")
    print(f"\nFirst few lines of generated SQL:")
    print("\n".join(sql.split("\n")[:10]))

if __name__ == "__main__":
    main()

