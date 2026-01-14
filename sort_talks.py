import json
from datetime import datetime

def sort_talks():
    file_path = 'data/talks.json'
    
    try:
        with open(file_path, 'r') as f:
            data = json.load(f)
        
        # Helper to get the latest event date for a talk
        def get_latest_date(talk):
            if not talk.get('events'):
                return datetime.min
            
            dates = []
            for event in talk['events']:
                try:
                    dates.append(datetime.strptime(event['date'], '%Y-%m-%d'))
                except ValueError:
                    continue # Skip invalid dates
            
            return max(dates) if dates else datetime.min

        # Sort talks by latest event date descending
        data['talks'].sort(key=get_latest_date, reverse=True)
        
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
            
        print("Successfully sorted talks.")
        
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    sort_talks()
