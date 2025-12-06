#!/usr/bin/env python3
"""
Script to translate monster_names.json from Chinese to English
"""

import json
import os
from pathlib import Path

# Translation dictionary (extracted from Go translations.go)
TRANSLATIONS = {
    # Basic Goblins
    "棒槌哥布林": "Club Goblin",
    "小刀哥布林": "Dagger Goblin",
    "弩箭哥布林": "Crossbow Goblin",
    "剑盾哥布林": "Sword & Shield Goblin",
    "巨斧哥布林": "Great Axe Goblin",
    "巫妖哥布林": "Necromancer Goblin",
    "巫师哥布林": "Wizard Goblin",
    "火焰哥布林": "Flame Goblin",
    "火焰哥布林巫师": "Flame Wizard Goblin",
    "棒槌火焰哥布林": "Club Flame Goblin",
    "短剑火焰哥布林": "Short Sword Flame Goblin",
    "风哥布林巫师": "Wind Wizard Goblin",
    "森哥布林巫师": "Forest Wizard Goblin",
    "飓风哥布林": "Hurricane Goblin",
    "飓风哥布林王": "Hurricane Goblin King",
    "飓风哥布林战士": "Hurricane Goblin Warrior",
    "短剑飓风哥布林": "Short Sword Hurricane Goblin",
    "法杖飓风哥布林": "Staff Hurricane Goblin",
    "虚蚀棒槌哥布林": "Void Corrupted Club Goblin",
    "虚蚀小刀哥布林": "Void Corrupted Dagger Goblin",
    "虚蚀弩箭哥布林": "Void Corrupted Crossbow Goblin",
    "虚蚀剑盾哥布林": "Void Corrupted Sword Goblin",
    "虚蚀巨斧哥布林": "Void Corrupted Great Axe Goblin",
    "虚蚀巫妖哥布林": "Void Corrupted Necromancer Goblin",
    "虚蚀火巫师哥布林": "Void Corrupted Flame Wizard Goblin",
    "虚蚀森巫师哥布林": "Void Corrupted Forest Wizard Goblin",
    "虚蚀风巫师哥布林": "Void Corrupted Wind Wizard Goblin",
    "被虚蚀的棒槌哥布林": "Void-Corrupted Club Goblin",
    "被虚蚀的小刀哥布林": "Void-Corrupted Dagger Goblin",
    "被虚蚀的弩箭哥布林": "Void-Corrupted Crossbow Goblin",
    "被虚蚀的剑盾哥布林": "Void-Corrupted Sword Goblin",
    "被虚蚀的巨斧哥布林": "Void-Corrupted Great Axe Goblin",
    "被虚蚀的巫师哥布林": "Void-Corrupted Wizard Goblin",
    "被虚蚀的巫妖哥布林": "Void-Corrupted Necromancer Goblin",
    "被虚蚀的火巫师哥布林": "Void-Corrupted Flame Wizard Goblin",
    "哥布林王": "Goblin King",
    "哥布林王·异种": "Goblin King Variant",
    "虚蚀·哥布林王": "Void Corrupted Goblin King",
    "虚蚀哥布林王": "Void Corrupted Goblin King",
    "哥布林护卫": "Goblin Guard",
    "法杖森林哥布林": "Staff Forest Goblin",
    "法杖火焰哥布林": "Staff Flame Goblin",
    "法杖飓风哥布林": "Staff Hurricane Goblin",
    "丛林哥布林战士": "Forest Goblin Warrior",

    # Ogres & Humanoids
    "寒霜食人魔": "Frost Ogre",
    "雷电食人魔": "Thunder Ogre",
    "火焰食人魔": "Flame Ogre",
    "雷电虚蚀的火焰食人魔": "Thunder Corrupted Flame Ogre",
    "雷电虚蚀的雷电食人魔": "Thunder Corrupted Thunder Ogre",
    "利奥雷乌斯": "Leoreaus",
    "伊戈雷乌斯": "Igoreus",
    "火焰兽人": "Flame Beast",
    "虚蚀·火焰兽人": "Void Corrupted Flame Beast",
    "[首领]火焰兽人": "[Boss] Flame Beast",

    # Mucks
    "棒槌姆克": "Club Muck",
    "大斧姆克": "Great Axe Muck",
    "弩手姆克": "Crossbow Muck",
    "大锤姆克": "Hammer Muck",
    "砍刀姆克": "Machete Muck",
    "疯狂双刀姆克": "Crazed Dual Blade Muck",
    "疯狂姆克弩手": "Crazed Muck Archer",
    "姆克兵长": "Muck Sergeant",
    "精英兵长姆克": "Elite Sergeant Muck",
    "姆克头目": "Muck Leader",
    "首领·姆克头目": "[Boss] Muck Leader",
    "姆克王": "Muck King",
    "首领·姆克王": "[Boss] Muck King",
    "姆克尖兵": "Elite Muck",
    "精英·姆克尖兵": "Elite Muck",
    "精英·姆克狂战士": "Elite Muck Berserker",
    "斥候姆克": "Scout Muck",

    # Beasts
    "野猪": "Wild Boar",
    "轰鸣野猪": "Thundering Boar",
    "暴烈野猪": "Savage Boar",
    "烈风野猪": "Storm Boar",
    "雷光野猪": "Thunder Boar",
    "高原野猪": "Highland Boar",
    "积雪野猪": "Snowy Boar",
    "赤炎野猪": "Crimson Flame Boar",
    "小猪": "Piglet",
    "冲锋野猪": "Charging Boar",
    "精英·变异肉山": "Elite Mutant Flesh Golem",
    "精英·烈风野猪": "Elite Storm Boar",
    "首领·野猪王": "[Boss] Boar King",
    "首领·赤炎野猪": "[Boss] Crimson Flame Boar",
    
    "角羊": "Horned Sheep",
    "凯撒角羊": "Caesar Sheep",
    "幽冥角羊": "Phantom Sheep",
    "翡翠角羊": "Jade Sheep",
    "燃烧角羊": "Burning Sheep",
    "梦魇角羊": "Nightmare Sheep",
    "精英·翡翠角羊": "Elite Jade Sheep",
    "精英·燃烧角羊": "Elite Burning Sheep",
    "首领·梦魇角羊": "[Boss] Nightmare Sheep",
    
    "蟹蛛": "Crab Spider",
    "小蟹蛛": "Small Crab Spider",
    "污染蟹蛛": "Polluted Crab Spider",
    "剧毒蟹蛛": "Deadly Poison Spider",
    "猛毒蟹蛛": "Venomous Spider",
    "蛰伏蟹蛛": "Lurking Spider",
    "幻妖蟹蛛": "Phantom Spider",
    "精英·污染蟹蛛": "Elite Polluted Spider",
    "首领·幻妖蟹蛛": "[Boss] Phantom Spider",
    
    "黄蜂": "Hornet",
    "致命黄蜂": "Deadly Hornet",
    "暴烈蜂": "Savage Bee",
    "变异蜂": "Mutant Bee",
    "剧毒蜂巢": "Deadly Poison Hive",
    "精英·变异蜂": "Elite Mutant Bee",
    "首领·变异蜂": "[Boss] Mutant Bee",
    
    "地狐": "Earth Fox",
    "赤玉地狐": "Crimson Jade Fox",
    "赤色飞沫": "Crimson Scatter",
    "蓝宝地狐": "Sapphire Fox",
    "变异地狐": "Mutant Fox",
    "精英·地狐": "Elite Earth Fox",
    "精英·高原地狐": "Elite Highland Fox",
    "精英·变异地狐": "Elite Mutant Fox",
    "精英·蓝宝地狐": "Elite Sapphire Fox",
    "首领·赤玉地狐": "[Boss] Crimson Jade Fox",
    "首领·嚎鸣地狐": "[Boss] Howling Fox",
    
    "熔岩肉山": "Lava Flesh Golem",
    "变异肉山": "Mutant Flesh Golem",
    "冰雪肉山": "Icy Flesh Golem",
    "首领·霸王肉山": "[Boss] Tyrant Flesh Golem",
    "多戈尔曼": "Dogolman",
    "首领·多戈尔曼": "[Boss] Dogolman",
    
    "岩石蜥蜴": "Stone Lizard",
    "蜥蜴": "Lizard",
    "蜥蜴人": "Lizardman",
    "蜥蜴人战士": "Lizardman Warrior",
    "蜥蜴人萨满": "Lizardman Shaman",
    "蜥蜴人猎手": "Lizardman Hunter",
    "蜥蜴人法师": "Lizardman Mage",
    "蜥蜴人王": "Lizardman King",
    "首领·蜥蜴人王": "[Boss] Lizardman King",

    # Birds
    "雄鹰": "Eagle",
    "鹞": "Hawk",
    "彩羽鹞": "Colorful Hawk",
    "皇后白羽鹞": "Queen White Hawk",
    "巨嘴鵎鵼": "Toucan",
    "紫冠鵎鵼": "Purple Crown Toucan",
    "精英·陆地鹰": "Elite Land Eagle",
    "精英·变异彩羽鹞": "Elite Mutant Hawk",
    "首领·陆地鹰": "[Boss] Land Eagle",
    "首领·皇后白羽鹞": "[Boss] Queen White Hawk",
    "首领·国王鵎鵼": "[Boss] King Toucan",

    # Bandits & Soldiers
    "山贼": "Bandit",
    "山贼射手": "Bandit Archer",
    "山贼斥候": "Bandit Scout",
    "山贼斧手": "Bandit Axeman",
    "山贼护卫": "Bandit Guard",
    "山贼打手": "Bandit Thug",
    "山贼首领": "Bandit Chief",
    "山贼头目": "Bandit Leader",
    "首领·山贼首领": "[Boss] Bandit Chief",
    "精英·山贼射手": "Elite Bandit Archer",
    "山贼护卫队长": "Bandit Guard Captain",
    "精英·山贼护卫队长": "Elite Bandit Guard Captain",
    "山贼头目战斧": "Bandit Leader War Axe",

    # Dark Legion
    "黯影军团·剑士": "Dark Legion Swordsman",
    "黯影军团·步枪兵": "Dark Legion Spearman",
    "黯影军团·盾兵": "Dark Legion Shield Bearer",
    "黯影军团·军官剑士": "Dark Legion Officer",
    "黯影军团士兵": "Dark Legion Soldier",
    "黯影军团男斥候": "Dark Legion Male Scout",
    "黯影军团男斥候精英": "Elite Dark Legion Male Scout",
    "黯影斥候": "Dark Scout",
    "黯影枪手": "Dark Gunner",
    "激进黯影剑士": "Radical Dark Swordsman",
    "激进黯影枪手": "Radical Dark Gunner",
    "激进黯影护卫": "Radical Dark Guard",
    "黯影剑士队长": "Dark Swordsman Captain",
    "黯影枪手队长": "Dark Gunner Captain",
    "虚蚀·黯影军团士兵": "Void Corrupted Dark Legion Soldier",
    "被虚蚀的黯影士兵": "Void-Corrupted Dark Soldier",
    "腐蚀·黯影士兵": "Corrupted Dark Soldier",

    # Machines/Constructs
    "保卫者04型": "Guardian 04",
    "捍卫者05型": "Defender 05",
    "捍卫者06型": "Defender 06",
    "哨兵01型": "Sentinel 01",
    "自爆兵02型": "Suicide Bomber 02",
    "自爆02型": "Suicide Unit 02",
    "自爆兵03型": "Suicide Bomber 03",
    "自爆机器人02型": "Suicide Robot 02",
    "自爆机器人08型": "Suicide Robot 08",
    "战斗机像03型": "Combat Automaton 03",
    "副战机像03型": "Support Automaton 03",
    "超级主战机像99型": "Super War Machine 99",
    "主战机像99型": "War Machine 99",
    "主战机残躯": "War Machine Wreck",
    "入侵者04型": "Invader 04",
    "追猎者02型": "Hunter 02",
    "湮灭堡垒01型": "Annihilation Fortress 01",
    "终焉毁灭者": "Final Destroyer",
    "边缘之刃": "Edge Blade",
    "灵魂精魄": "Soul Specter",
    "激光炮": "Laser Cannon",
    "精锐捍卫者06型": "Elite Defender 06",
    "精锐哨兵01型": "Elite Sentinel 01",
    "机像核心": "Automaton Core",

    # Characters/NPCs
    "蒂娜": "Tina",
    "艾露娜": "Alena",
    "奥尔维拉": "Oliveira",
    "塔塔": "Tata",
    "杰拉德": "Gerard",
    "罗罗拉": "Rorola",
    "迷失的蒂娜": "Lost Tina",
    "蒂娜·虚蚀心像": "Tina - Void Echo",
    "蒂娜·虚蚀心像幻影": "Tina - Void Echo Phantom",

    # Boss/Elite prefix patterns (must be checked after exact matches)
    "首领·": "[Boss] ",
    "精英·": "Elite: ",
    "[首领]": "[Boss] ",

    # Kanimals
    "卡尼曼战士": "Kanimal Warrior",
    "卡尼曼猎人": "Kanimal Hunter",
    "卡尼曼射手": "Kanimal Archer",
    "卡尼曼巫师": "Kanimal Wizard",
    "卡尼曼巫毒战士": "Kanimal Voodoo Warrior",
    "卡尼曼刺杀者": "Kanimal Assassin",
    "卡尼曼斥候": "Kanimal Scout",
    "卡尼曼高阶猎手": "Kanimal Elite Hunter",
    "卡尼曼游魂": "Kanimal Spirit",
    "卡尼曼亡灵战士": "Kanimal Undead Warrior",
    "卡尼曼亡灵猎人": "Kanimal Undead Hunter",
    "卡尼曼亡灵射手": "Kanimal Undead Archer",
    "卡尼曼亡灵自爆怪": "Kanimal Undead Suicide Unit",
    "卡尼曼亡灵战神": "[Boss] Kanimal Undead God",
    "卡尼曼巨人": "Kanimal Giant",
    "卡尼曼缚魂者": "Kanimal Soul Binder",
    "卡尼曼猎魂者": "Kanimal Soul Hunter",
    "精英·卡尼曼射手": "Elite Kanimal Archer",
    "精英·卡尼曼战士": "Elite Kanimal Warrior",

    # Ascantrians
    "阿斯特里斯士兵": "Ascantrian Soldier",
    "阿斯特里斯护卫": "Ascantrian Guard",
    "阿斯特里斯士兵A": "Ascantrian Soldier A",
    "阿斯特里斯士兵B": "Ascantrian Soldier B",
    "阿斯特里斯士兵C": "Ascantrian Soldier C",

    # Banhalt
    "班哈尔特剑士": "Banhalt Swordsman",
    "班哈尔特步枪兵": "Banhalt Spearman",
    "班哈尔特盾兵": "Banhalt Shield Bearer",
    "班哈尔特长矛兵": "Banhalt Pikeman",
    "班哈尔特坦克": "Banhalt Tank",
    "班哈尔特巨型机像": "Banhalt Giant Automaton",
    "班哈尔特盾手": "Banhalt Shield Warrior",
    "班哈尔特守御士": "Banhalt Defender",
    "班哈尔特步枪手": "Banhalt Spear Warrior",
    "班哈尔特长矛手": "Banhalt Pike Warrior",
    "班哈尔特光击士": "Banhalt Light Warrior",
    "班哈尔特战争机器": "Banhalt War Machine",
    "精英·班哈尔特剑士": "Elite Banhalt Swordsman",
    "精英·班哈尔特护卫": "Elite Banhalt Guard",

    # Misc
    "木桩": "Training Dummy",
    "木人": "Wooden Golem",
    "测试": "Test",
    "变身专用": "Transform Only",
    "灾厄傀儡": "Calamity Puppet",
    "灾厄之主": "Lord of Calamity",
    "虫菇": "Fungal Bug",
}


def translate_name(chinese_name):
    """Translate a single monster name from Chinese to English"""
    
    # Try exact match first
    if chinese_name in TRANSLATIONS:
        return TRANSLATIONS[chinese_name]
    
    # Try partial matches for complex names (sorted by length descending for best match)
    sorted_translations = sorted(TRANSLATIONS.items(), key=lambda x: len(x[0]), reverse=True)
    for chinese, english in sorted_translations:
        if len(chinese) > 2 and chinese in chinese_name:
            # Replace the Chinese part with English
            result = chinese_name.replace(chinese, english, 1)
            return result
    
    # If no translation found, return original name
    return chinese_name


def translate_json_file(input_path, output_path):
    """Translate all monster names in a JSON file"""
    
    print(f"Reading from: {input_path}")
    
    # Read the JSON file
    with open(input_path, 'r', encoding='utf-8') as f:
        data = json.load(f)
    
    # Translate all names
    translated = {}
    untranslated_count = 0
    translated_count = 0
    
    for key, chinese_name in data.items():
        english_name = translate_name(chinese_name)
        translated[key] = english_name
        
        if english_name == chinese_name:
            untranslated_count += 1
            print(f"⚠️  {key}: {chinese_name} (NO TRANSLATION)")
        else:
            translated_count += 1
            if translated_count <= 10:  # Show first 10 translations
                print(f"✓ {key}: {chinese_name} → {english_name}")
    
    # Write the translated JSON file
    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(translated, f, ensure_ascii=False, indent=2)
    
    print(f"\n" + "="*80)
    print(f"Translation complete!")
    print(f"Total monsters: {len(translated)}")
    print(f"Translated: {translated_count}")
    print(f"Not translated: {untranslated_count}")
    print(f"Output saved to: {output_path}")


if __name__ == "__main__":
    # File paths
    input_file = Path(__file__).parent / "global" / "monster_names.json"
    output_file = Path(__file__).parent / "global" / "monster_names_en.json"
    
    if not input_file.exists():
        print(f"ERROR: Input file not found: {input_file}")
        exit(1)
    
    translate_json_file(str(input_file), str(output_file))
