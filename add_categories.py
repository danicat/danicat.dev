import json

def add_categories():
    file_path = 'data/talks.json'
    
    try:
        with open(file_path, 'r') as f:
            data = json.load(f)
        
        for talk in data['talks']:
            # Default category
            category = "AI & Development"
            
            # Simple heuristic to assign categories if needed, otherwise default
            # For now, assign all to "AI & Development" as requested/safe default
            # consistently with the blog's "AI & Development" category.
            
            talk['categories'] = [category]
        
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
            
        print("Successfully added categories to talks.")
        
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    add_categories()
