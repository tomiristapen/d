import os
import re

import pandas as pd


BASE_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

FOOD_PATH = os.path.join(BASE_DIR, "raw_data", "food.csv")
NUTRIENT_PATH = os.path.join(BASE_DIR, "raw_data", "nutrient.csv")
FOOD_NUTRIENT_PATH = os.path.join(BASE_DIR, "raw_data", "food_nutrient.csv")

PRODUCTS_OUTPUT_PATH = os.path.join(BASE_DIR, "output", "products_clean.csv")
ALIASES_OUTPUT_PATH = os.path.join(BASE_DIR, "output", "products_aliases.csv")
LEGACY_OUTPUT_PATH = os.path.join(BASE_DIR, "output", "products_base_import.csv")

SOURCE_SYSTEM = "usda"

DATA_TYPE_PRIORITY = {
    "foundation_food": 0,
    "agricultural_acquisition": 1,
    "sample_food": 2,
    "market_acquisition": 3,
}

EXCLUDED_DATA_TYPES = {"sub_sample_food"}
BAD_NAME_KEYWORDS = ("restaurant", "fast food")
IRREGULAR_SINGULARS = {
    "apples": "apple",
    "bananas": "banana",
    "beans": "bean",
    "carrots": "carrot",
    "cookies": "cookie",
    "figs": "fig",
    "melons": "melon",
    "nectarines": "nectarine",
    "olives": "olive",
    "onions": "onion",
    "oranges": "orange",
    "peaches": "peach",
    "pears": "pear",
    "pickles": "pickle",
    "seeds": "seed",
    "strawberries": "strawberry",
    "sugars": "sugar",
    "tomatoes": "tomato",
}
NOISE_SEGMENTS = (
    "with added vitamin",
    "includes foods for usda",
    "food distribution program",
    "fluid",
    "meat only",
    "separable lean only",
    "trimmed to",
    "choice",
    "select",
    "regular pack",
    "pre-cooked",
    "unprepared",
    "prepared",
    "commercially prepared",
    "commercial",
    "ready-to-serve",
    "water added",
    "vitamin d fortified",
    "pre-packaged",
    "deli meat",
    "unheated",
    "pasteurized",
    "boneless",
    "skinless",
    "without salt",
    "with salt added",
)
UNCOUNTABLE_WORDS = {
    "bread",
    "broccoli",
    "cheese",
    "fish",
    "haddock",
    "ham",
    "hummus",
    "kale",
    "kiwifruit",
    "milk",
    "mustard",
    "pollock",
    "rice",
    "salsa",
    "sauce",
    "sugar",
    "tuna",
    "yogurt",
}
PRODUCE_CORES = {
    "apple",
    "banana",
    "bean",
    "broccoli",
    "carrot",
    "fig",
    "grapefruit",
    "kale",
    "kiwifruit",
    "lettuce",
    "melon",
    "nectarine",
    "olive",
    "onion",
    "orange",
    "peach",
    "pear",
    "pickle",
    "strawberry",
    "tomato",
}
PRODUCE_VARIETIES = {
    "bartlett",
    "cantaloupe",
    "cos",
    "dill",
    "fuji",
    "gala",
    "granny smith",
    "grape",
    "green",
    "honeycrisp",
    "kosher dill",
    "manzanilla",
    "navel",
    "red",
    "red delicious",
    "romaine",
    "white",
    "yellow",
}
GENERIC_ALLOWED_MODIFIERS = {
    "all-purpose",
    "american",
    "almonds",
    "breakfast",
    "brown",
    "cheddar",
    "chorizo",
    "coconut",
    "corn",
    "cottage",
    "dry roasted",
    "granulated",
    "greek",
    "ground",
    "italian",
    "low moisture",
    "mozzarella",
    "nonfat",
    "oatmeal",
    "parmesan",
    "part-skim",
    "peanut",
    "plain",
    "pork",
    "queso seco",
    "reduced fat",
    "rice",
    "smooth",
    "soy",
    "spaghetti/marinara",
    "sunflower seed kernels",
    "swiss",
    "table",
    "turkey",
    "whole wheat",
    "whole-wheat",
    "white",
    "wheat",
    "yellow",
}
PREPARATION_KEYWORDS = {
    "canned": "canned",
    "dried": "dried",
    "frozen": "frozen",
    "boiled": "boiled",
    "braised": "braised",
    "grilled": "grilled",
    "roasted": "roasted",
    "fried": "fried",
    "cooked": "cooked",
}
MEAT_CUT_PATTERNS = (
    "porterhouse steak",
    "t-bone steak",
    "top loin steak",
    "eye of round roast",
    "top round roast",
    "tenderloin roast",
    "tenderloin",
    "drumstick",
    "breast",
    "thigh",
    "links",
    "ground",
    "sausage",
    "roast",
    "loin",
    "steak",
)
FISH_SPECIES = ("haddock", "pollock", "tuna")


def normalize_spaces(value):
    return re.sub(r"\s+", " ", value).strip()


def clean_source_name(value):
    if pd.isna(value):
        return ""

    text = str(value).lower()
    text = text.replace("&", " and ")
    text = re.sub(r"\([^)]*\)", "", text)
    text = text.replace("'", "")
    text = normalize_spaces(text)
    return text


def singularize(word):
    word = normalize_spaces(word)
    if " " in word:
        parts = word.split(" ")
        parts[-1] = singularize(parts[-1])
        return " ".join(parts)
    if word in IRREGULAR_SINGULARS:
        return IRREGULAR_SINGULARS[word]
    if not word or word in UNCOUNTABLE_WORDS:
        return word
    if word.endswith("ies") and len(word) > 3:
        return word[:-3] + "y"
    if word.endswith("oes") and len(word) > 3:
        return word[:-2]
    if word.endswith("s") and not word.endswith(("ss", "us")):
        return word[:-1]
    return word


def is_noise_segment(segment):
    lowered = normalize_spaces(segment.lower())
    if not lowered:
        return True
    return any(noise in lowered for noise in NOISE_SEGMENTS)


def is_bad_name(text):
    lowered = text.lower()
    return any(word in lowered for word in BAD_NAME_KEYWORDS)


def extract_preparation(modifiers):
    tags = []
    text = ", ".join(modifiers)
    for keyword, label in PREPARATION_KEYWORDS.items():
        if keyword in text and label not in tags:
            tags.append(label)
    return tags


def first_meaningful_modifier(modifiers, allowed=None):
    for modifier in modifiers:
        cleaned = normalize_spaces(modifier)
        if not cleaned or is_noise_segment(cleaned):
            continue
        if allowed is not None and cleaned not in allowed:
            continue
        return cleaned
    return ""


def match_modifier_keyword(modifiers, allowed):
    text = ", ".join(normalize_spaces(modifier) for modifier in modifiers).lower()
    for option in sorted(allowed, key=len, reverse=True):
        if option in text:
            return option
    return ""


def canonicalize_milk(modifiers):
    text = ", ".join(modifiers)
    if "nonfat" in text or "skim" in text or "fat free" in text:
        return "skim milk"
    if "whole" in text or "3.25%" in text:
        return "whole milk"
    match = re.search(r"(\d+(?:\.\d+)?)%\s*milkfat", text)
    if match:
        value = match.group(1).rstrip("0").rstrip(".")
        return f"milk {value}%"
    match = re.search(r"milk,\s*(\d+(?:\.\d+)?)%", "milk, " + text)
    if match:
        value = match.group(1).rstrip("0").rstrip(".")
        return f"milk {value}%"
    return "milk"


def canonicalize_egg(core, modifiers):
    text = ", ".join(modifiers)
    if "white" in text:
        name = "egg white"
    elif "yolk" in text:
        name = "egg yolk"
    elif core == "eggs" and "whole" in text:
        name = "egg"
    else:
        name = "egg"

    if "dried" in text:
        return f"dried {name}"
    return name


def canonicalize_produce(base, modifiers):
    variety = match_modifier_keyword(modifiers, PRODUCE_VARIETIES)
    if variety:
        variety = singularize(variety)
        name = f"{variety} {base}"
    else:
        name = base

    preparation = extract_preparation(modifiers)
    if "canned" in preparation:
        name = f"{name} canned"
    elif "dried" in preparation:
        name = f"{name} dried"
    elif "frozen" in preparation:
        name = f"{name} frozen"

    return normalize_spaces(name)


def canonicalize_bread(modifiers):
    modifier = match_modifier_keyword(modifiers, {"whole wheat", "whole-wheat", "white"})
    if modifier in {"whole wheat", "whole-wheat"}:
        return "whole wheat bread"
    if modifier:
        return f"{modifier} bread"
    return "bread"


def canonicalize_cheese(modifiers):
    modifier = match_modifier_keyword(
        modifiers,
        {"american", "cheddar", "cottage", "mozzarella", "parmesan", "queso seco", "ricotta", "swiss"},
    )
    if modifier == "cottage":
        return "cottage cheese"
    if modifier == "queso seco":
        return "queso seco cheese"
    if modifier:
        return f"{modifier} cheese"
    return "cheese"


def canonicalize_sauce(modifiers):
    text = ", ".join(modifiers)
    if "salsa" in text:
        return "salsa"
    if "spaghetti/marinara" in text:
        return "marinara sauce"
    return "sauce"


def canonicalize_sugar(modifiers):
    if match_modifier_keyword(modifiers, {"granulated"}):
        return "granulated sugar"
    return "sugar"


def canonicalize_nuts(modifiers):
    modifier = match_modifier_keyword(modifiers, {"almonds"})
    if modifier:
        return singularize(modifier)
    return "nuts"


def canonicalize_seeds(modifiers):
    if match_modifier_keyword(modifiers, {"sunflower seed kernels"}):
        return "sunflower seeds"
    return "seeds"


def canonicalize_flour(modifiers):
    modifier = match_modifier_keyword(
        modifiers,
        {"all-purpose", "bread", "corn", "rice", "soy", "whole wheat", "wheat"},
    )
    if modifier == "all-purpose":
        return "all-purpose flour"
    if modifier == "whole wheat":
        return "whole wheat flour"
    if modifier:
        return f"{modifier} flour"
    return "flour"


def canonicalize_meat(base, modifiers):
    text = ", ".join(modifiers)
    cut = ""
    for pattern in MEAT_CUT_PATTERNS:
        if pattern in text:
            cut = pattern
            break

    if base == "fish":
        species = match_modifier_keyword(modifiers, set(FISH_SPECIES))
        name = species or "fish"
    elif cut:
        name = f"{base} {cut}"
    else:
        modifier = match_modifier_keyword(modifiers, {"ground", "italian", "turkey", "beef", "pork"})
        name = f"{modifier} {base}" if modifier else base

    preparation = extract_preparation(modifiers)
    for tag in preparation:
        if tag not in {"cooked"}:
            name = f"{name} {tag}"
            break
    return normalize_spaces(name)


def canonicalize_generic(base, modifiers):
    modifier = match_modifier_keyword(modifiers, GENERIC_ALLOWED_MODIFIERS)
    if modifier:
        name = f"{modifier} {base}"
    else:
        name = base

    preparation = extract_preparation(modifiers)
    if preparation and base not in {"sauce", "salsa"}:
        first_tag = preparation[0]
        if first_tag not in {"cooked"}:
            name = f"{name} {first_tag}"
    return normalize_spaces(name)


def canonicalize_name(description):
    source_name = clean_source_name(description)
    if not source_name or is_bad_name(source_name):
        return ""

    segments = [normalize_spaces(part) for part in source_name.split(",") if normalize_spaces(part)]
    if not segments:
        return ""

    core = segments[0]
    modifiers = segments[1:]
    base = singularize(core)

    if base == "milk":
        return canonicalize_milk(modifiers)
    if base == "egg":
        return canonicalize_egg(core, modifiers)
    if base in PRODUCE_CORES:
        return canonicalize_produce(base, modifiers)
    if base == "bread":
        return canonicalize_bread(modifiers)
    if base == "cheese":
        return canonicalize_cheese(modifiers)
    if base == "flour":
        return canonicalize_flour(modifiers)
    if base == "nut":
        return canonicalize_nuts(modifiers)
    if base == "sauce":
        return canonicalize_sauce(modifiers)
    if base == "seed":
        return canonicalize_seeds(modifiers)
    if base == "sugar":
        return canonicalize_sugar(modifiers)
    if base in {"beef", "chicken", "fish", "ham", "pork", "turkey"}:
        return canonicalize_meat(base, modifiers)
    return canonicalize_generic(base, modifiers)


def pick_nutrient_id(frame, nutrient_nbr, allowed_names):
    matched = frame[
        frame["nutrient_nbr"].astype(str).eq(str(nutrient_nbr))
        | frame["name"].isin(allowed_names)
    ].copy()
    if matched.empty:
        raise RuntimeError(f"Could not find nutrient id for {allowed_names}")

    if nutrient_nbr == 208:
        kcal = matched[matched["unit_name"].astype(str).str.upper() == "KCAL"]
        if not kcal.empty:
            return int(kcal.iloc[0]["id"]), "KCAL"
    row = matched.iloc[0]
    return int(row["id"]), str(row["unit_name"]).upper()


def build_products():
    food = pd.read_csv(FOOD_PATH)
    nutrient = pd.read_csv(NUTRIENT_PATH)
    food_nutrient = pd.read_csv(FOOD_NUTRIENT_PATH, low_memory=False)

    calories_id, calories_unit = pick_nutrient_id(nutrient, 208, {"Energy"})
    protein_id, _ = pick_nutrient_id(nutrient, 203, {"Protein"})
    fat_id, _ = pick_nutrient_id(nutrient, 204, {"Total lipid (fat)"})
    carbs_id, _ = pick_nutrient_id(nutrient, 205, {"Carbohydrate, by difference"})

    needed_ids = {
        calories_id: "calories",
        protein_id: "protein",
        fat_id: "fat",
        carbs_id: "carbs",
    }

    filtered = food_nutrient[food_nutrient["nutrient_id"].isin(needed_ids.keys())].copy()

    pivot = filtered.pivot_table(
        index="fdc_id",
        columns="nutrient_id",
        values="amount",
        aggfunc="first",
    ).reset_index()
    pivot = pivot.rename(columns=needed_ids)

    if calories_unit == "KJ":
        pivot["calories"] = pivot["calories"] / 4.184

    result = pivot.merge(food, on="fdc_id", how="inner")
    result = result[~result["data_type"].isin(EXCLUDED_DATA_TYPES)].copy()

    result["source_name"] = result["description"].apply(clean_source_name)
    result["name"] = result["description"].apply(canonicalize_name)
    result["normalized_name"] = result["name"].apply(clean_source_name)
    result["data_type_priority"] = result["data_type"].map(DATA_TYPE_PRIORITY).fillna(999)

    result = result[result["normalized_name"] != ""].copy()
    result = result.dropna(subset=["calories"]).copy()
    result["protein"] = result["protein"].fillna(0.0)
    result["fat"] = result["fat"].fillna(0.0)
    result["carbs"] = result["carbs"].fillna(0.0)

    result = result.sort_values(["normalized_name", "data_type_priority", "fdc_id"])

    products = (
        result.groupby("normalized_name", as_index=False)
        .agg(
            name=("name", "first"),
            source_name=("source_name", "first"),
            source_ref=("fdc_id", "first"),
            food_category_id=("food_category_id", "first"),
            source_count=("fdc_id", "count"),
            data_types=("data_type", lambda s: ", ".join(sorted(set(map(str, s))))),
            calories=("calories", "mean"),
            protein=("protein", "mean"),
            fat=("fat", "mean"),
            carbs=("carbs", "mean"),
        )
        .copy()
    )

    products["source_system"] = SOURCE_SYSTEM
    products["calories"] = products["calories"].round(2)
    products["protein"] = products["protein"].round(2)
    products["fat"] = products["fat"].round(2)
    products["carbs"] = products["carbs"].round(2)

    products = products[
        [
            "name",
            "normalized_name",
            "source_name",
            "source_system",
            "source_ref",
            "source_count",
            "food_category_id",
            "data_types",
            "calories",
            "protein",
            "fat",
            "carbs",
        ]
    ].sort_values("name")

    aliases = build_aliases(result, products)

    return products, aliases


def build_aliases(source_rows, products):
    aliases = []

    for _, row in products.iterrows():
        aliases.extend(build_alias_records(row["name"], row["normalized_name"], "canonical"))

    alias_candidates = (
        source_rows[["name", "source_name"]]
        .drop_duplicates()
        .rename(columns={"source_name": "alias"})
        .copy()
    )
    alias_candidates["alias_normalized"] = alias_candidates["alias"].apply(clean_source_name)
    alias_candidates["alias_type"] = "source_name"

    for _, row in alias_candidates.iterrows():
        if not row["alias_normalized"] or row["alias_normalized"] == clean_source_name(row["name"]):
            continue
        aliases.append(
            {
                "product_name": row["name"],
                "alias": row["alias"],
                "alias_normalized": row["alias_normalized"],
                "alias_type": row["alias_type"],
            }
        )

    aliases_df = pd.DataFrame(aliases).drop_duplicates().sort_values(["product_name", "alias_type", "alias"])
    return aliases_df


def build_alias_records(product_name, normalized_name, alias_type):
    records = [
        {
            "product_name": product_name,
            "alias": product_name,
            "alias_normalized": normalized_name,
            "alias_type": alias_type,
        }
    ]
    for alias in sorted(derived_aliases(product_name)):
        alias_normalized = clean_source_name(alias)
        if not alias_normalized or alias_normalized == normalized_name:
            continue
        records.append(
            {
                "product_name": product_name,
                "alias": alias,
                "alias_normalized": alias_normalized,
                "alias_type": "derived",
            }
        )
    return records


def derived_aliases(product_name):
    name = clean_source_name(product_name)
    aliases = set()
    if not name:
        return aliases

    if name.endswith(" apple") and name != "apple":
        aliases.add("apple")
    if name.endswith(" pear") and name != "pear":
        aliases.add("pear")
    if name.endswith(" peach") and name != "peach":
        aliases.add("peach")
    if name.endswith(" tomato") and name != "tomato":
        aliases.add("tomato")
    if name.endswith(" lettuce") and name != "lettuce":
        aliases.add("lettuce")
    if name.endswith(" onion") and name != "onion":
        aliases.add("onion")
    if name.endswith(" milk") or name.startswith("milk "):
        aliases.add("milk")
    if name.endswith(" cheese") and name != "cheese":
        aliases.add("cheese")
    if name.endswith(" flour") and name != "flour":
        aliases.add("flour")

    for suffix in (" braised", " grilled", " roasted", " fried", " frozen", " canned"):
        if name.endswith(suffix):
            aliases.add(name[: -len(suffix)])

    return aliases


def main():
    os.makedirs(os.path.dirname(PRODUCTS_OUTPUT_PATH), exist_ok=True)

    products, aliases = build_products()
    products.to_csv(PRODUCTS_OUTPUT_PATH, index=False)
    aliases.to_csv(ALIASES_OUTPUT_PATH, index=False)
    products[["name", "calories", "protein", "fat", "carbs"]].to_csv(LEGACY_OUTPUT_PATH, index=False)

    print("Saved products:", len(products), "->", PRODUCTS_OUTPUT_PATH)
    print("Saved aliases:", len(aliases), "->", ALIASES_OUTPUT_PATH)
    print("Saved legacy import:", len(products), "->", LEGACY_OUTPUT_PATH)


if __name__ == "__main__":
    main()
