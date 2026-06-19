-- Flag menu items recommended by the chef

ALTER TABLE menus ADD COLUMN IF NOT EXISTS is_chefs_choice BOOLEAN DEFAULT FALSE;
