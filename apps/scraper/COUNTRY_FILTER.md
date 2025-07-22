# Country Filter Feature

The QuakeWatch Scraper now supports filtering earthquakes by country. This feature allows you to collect earthquake data for specific countries or regions.

## Usage

### Basic Country Filter

```bash
# Collect earthquakes in Japan (last 30 days, all magnitudes)
./bin/quakewatch-scraper earthquakes country --country "Japan"

# Collect earthquakes in California, USA
./bin/quakewatch-scraper earthquakes country --country "California"

# Collect earthquakes in Australia with magnitude 4.0+
./bin/quakewatch-scraper earthquakes country --country "Australia" --min-mag 4.0
```

### Advanced Filtering

```bash
# Collect earthquakes in Chile with specific time range and magnitude
./bin/quakewatch-scraper earthquakes country \
  --country "Chile" \
  --start "2024-01-01" \
  --end "2024-12-31" \
  --min-mag 5.0 \
  --max-mag 8.0 \
  --limit 100

# Collect recent significant earthquakes in Turkey
./bin/quakewatch-scraper earthquakes country \
  --country "Turkey" \
  --min-mag 4.5 \
  --limit 50 \
  --filename "turkey_significant_2024"
```

## Supported Countries

The country filter supports major countries and their common variations:

### North America
- **USA/United States**: Includes all US states (California, Texas, Alaska, etc.)
- **Canada**: Includes major cities and provinces
- **Mexico**: Includes major cities and states

### Asia
- **Japan**: Includes major islands and cities (Honshu, Hokkaido, Tokyo, Osaka, etc.)
- **China**: Includes major cities and regions
- **Indonesia**: Includes major islands and cities (Java, Sumatra, Bali, Jakarta, etc.)
- **Philippines**: Includes major islands and cities (Luzon, Visayas, Mindanao, Manila, etc.)
- **Turkey**: Includes major cities and regions

### Oceania
- **Australia**: Includes major cities and states
- **New Zealand**: Includes North Island, South Island, and major cities

### South America
- **Chile**: Includes major cities and regions
- **Peru**: Includes major cities
- **Ecuador**: Includes major cities
- **Colombia**: Includes major cities

### Europe
- **Italy**: Includes major cities and regions
- **Greece**: Includes major cities and islands

## How It Works

The country filter works by:

1. **Fetching Data**: First, it fetches earthquake data from the USGS API based on your time range and magnitude filters
2. **Post-Processing**: Then it filters the results by checking if the earthquake's `place` field contains the specified country name
3. **Smart Matching**: The filter uses case-insensitive matching and includes common variations for major countries

### Example Place Fields

USGS earthquake data includes place descriptions like:
- "10km ENE of Tokyo, Japan"
- "5km SW of Los Angeles, CA"
- "Near the coast of Chile"
- "Central Turkey"

The filter will match these based on the country name you specify.

## Command Options

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `--country` | Country name to filter by | - | Yes |
| `--start` | Start time (YYYY-MM-DD) | 30 days ago | No |
| `--end` | End time (YYYY-MM-DD) | Current time | No |
| `--min-mag` | Minimum magnitude | 0.0 | No |
| `--max-mag` | Maximum magnitude | 10.0 | No |
| `--limit` | Maximum number of records | 1000 | No |
| `--filename` | Custom filename (without extension) | Auto-generated | No |

## Examples

### Recent Earthquakes in Japan
```bash
./bin/quakewatch-scraper earthquakes country --country "Japan" --limit 20
```

### Significant Earthquakes in California
```bash
./bin/quakewatch-scraper earthquakes country \
  --country "California" \
  --min-mag 4.5 \
  --start "2024-01-01" \
  --end "2024-12-31"
```

### All Earthquakes in Turkey (Last 30 Days)
```bash
./bin/quakewatch-scraper earthquakes country --country "Turkey"
```

## Limitations

1. **Place Field Dependency**: The filter relies on the accuracy of the `place` field in USGS data
2. **Language Variations**: Some countries may have multiple name variations
3. **Geographic Boundaries**: The filter uses text matching, not precise geographic boundaries
4. **API Limits**: Large time ranges may hit USGS API limits

## Tips

- Use specific country names for better results
- For US states, you can use the state name directly (e.g., "California", "Texas")
- Combine with magnitude filters to focus on significant events
- Use reasonable time ranges to avoid API timeouts
- The filter is case-insensitive, so "japan", "Japan", and "JAPAN" all work the same

## Output

The filtered earthquakes are saved to JSON files in the configured output directory, following the same format as other earthquake data collection commands. 