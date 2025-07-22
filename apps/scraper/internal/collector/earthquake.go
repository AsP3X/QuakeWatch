package collector

import (
	"fmt"
	"strings"
	"time"

	"quakewatch-scraper/internal/api"
	"quakewatch-scraper/internal/models"
	"quakewatch-scraper/internal/storage"
)

// EarthquakeCollector handles collecting earthquake data
type EarthquakeCollector struct {
	usgsClient *api.USGSClient
	storage    *storage.JSONStorage
}

// NewEarthquakeCollector creates a new earthquake collector
func NewEarthquakeCollector(usgsClient *api.USGSClient, storage *storage.JSONStorage) *EarthquakeCollector {
	return &EarthquakeCollector{
		usgsClient: usgsClient,
		storage:    storage,
	}
}

// CollectRecent collects recent earthquakes (last hour)
func (c *EarthquakeCollector) CollectRecent(limit int, filename string) error {
	fmt.Printf("Collecting recent earthquakes (last hour, limit: %d)...\n", limit)

	earthquakes, err := c.usgsClient.GetRecentEarthquakes(limit)
	if err != nil {
		return fmt.Errorf("failed to fetch recent earthquakes: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectByTimeRange collects earthquakes within a specific time range
func (c *EarthquakeCollector) CollectByTimeRange(startTime, endTime time.Time, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes from %s to %s (limit: %d)...\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByTimeRange(startTime, endTime, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by time range: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectByMagnitude collects earthquakes within a magnitude range
func (c *EarthquakeCollector) CollectByMagnitude(minMag, maxMag float64, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes with magnitude %.1f to %.1f (limit: %d)...\n", minMag, maxMag, limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByMagnitude(minMag, maxMag, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by magnitude: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectSignificant collects significant earthquakes (M4.5+)
func (c *EarthquakeCollector) CollectSignificant(startTime, endTime time.Time, limit int, filename string) error {
	fmt.Printf("Collecting significant earthquakes (M4.5+) from %s to %s (limit: %d)...\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		limit)

	earthquakes, err := c.usgsClient.GetSignificantEarthquakes(startTime, endTime, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch significant earthquakes: %w", err)
	}

	fmt.Printf("Found %d significant earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved significant earthquakes to %s\n", filename)
	return nil
}

// CollectByRegion collects earthquakes within a geographic region
func (c *EarthquakeCollector) CollectByRegion(minLat, maxLat, minLon, maxLon float64, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes in region (%.2f,%.2f) to (%.2f,%.2f) (limit: %d)...\n",
		minLat, minLon, maxLat, maxLon, limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByRegion(minLat, maxLat, minLon, maxLon, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by region: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(earthquakes, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectByCountry collects earthquakes filtered by country name
func (c *EarthquakeCollector) CollectByCountry(country string, startTime, endTime time.Time, minMag, maxMag float64, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes in %s from %s to %s (magnitude %.1f-%.1f, limit: %d)...\n",
		country,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		minMag, maxMag, limit)

	// First, fetch earthquakes by time range and magnitude
	earthquakes, err := c.usgsClient.GetEarthquakesByTimeRangeAndMagnitude(startTime, endTime, minMag, maxMag, limit*2) // Fetch more to account for filtering
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes: %w", err)
	}

	// Filter earthquakes by country
	var filteredEarthquakes []models.Earthquake
	for _, eq := range earthquakes.Features {
		if containsCountry(eq.Properties.Place, country) {
			filteredEarthquakes = append(filteredEarthquakes, eq)
		}
	}

	// Limit the results
	if len(filteredEarthquakes) > limit {
		filteredEarthquakes = filteredEarthquakes[:limit]
	}

	// Create a new response with filtered earthquakes
	filteredResponse := &models.USGSResponse{
		Type:     earthquakes.Type,
		Metadata: earthquakes.Metadata,
		Features: filteredEarthquakes,
	}

	// Update metadata count
	filteredResponse.Metadata.Count = len(filteredEarthquakes)

	fmt.Printf("Found %d earthquakes in %s\n", len(filteredEarthquakes), country)

	if err := c.storage.SaveEarthquakes(filteredResponse, filename); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectRecentData collects recent earthquakes and returns the data without saving
func (c *EarthquakeCollector) CollectRecentData(limit int) (*models.USGSResponse, error) {
	fmt.Printf("Collecting recent earthquakes (last hour, limit: %d)...\n", limit)

	earthquakes, err := c.usgsClient.GetRecentEarthquakes(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recent earthquakes: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))
	return earthquakes, nil
}

// CollectByTimeRangeData collects earthquakes within a specific time range and returns the data without saving
func (c *EarthquakeCollector) CollectByTimeRangeData(startTime, endTime time.Time, limit int) (*models.USGSResponse, error) {
	fmt.Printf("Collecting earthquakes from %s to %s (limit: %d)...\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByTimeRange(startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch earthquakes by time range: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))
	return earthquakes, nil
}

// CollectByMagnitudeData collects earthquakes within a magnitude range and returns the data without saving
func (c *EarthquakeCollector) CollectByMagnitudeData(minMag, maxMag float64, limit int) (*models.USGSResponse, error) {
	fmt.Printf("Collecting earthquakes with magnitude %.1f to %.1f (limit: %d)...\n", minMag, maxMag, limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByMagnitude(minMag, maxMag, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch earthquakes by magnitude: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))
	return earthquakes, nil
}

// CollectSignificantData collects significant earthquakes and returns the data without saving
func (c *EarthquakeCollector) CollectSignificantData(startTime, endTime time.Time, limit int) (*models.USGSResponse, error) {
	fmt.Printf("Collecting significant earthquakes (M4.5+) from %s to %s (limit: %d)...\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		limit)

	earthquakes, err := c.usgsClient.GetSignificantEarthquakes(startTime, endTime, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch significant earthquakes: %w", err)
	}

	fmt.Printf("Found %d significant earthquakes\n", len(earthquakes.Features))
	return earthquakes, nil
}

// CollectByRegionData collects earthquakes within a geographic region and returns the data without saving
func (c *EarthquakeCollector) CollectByRegionData(minLat, maxLat, minLon, maxLon float64, limit int) (*models.USGSResponse, error) {
	fmt.Printf("Collecting earthquakes in region (%.2f,%.2f) to (%.2f,%.2f) (limit: %d)...\n",
		minLat, minLon, maxLat, maxLon, limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByRegion(minLat, maxLat, minLon, maxLon, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch earthquakes by region: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))
	return earthquakes, nil
}

// CollectByCountryData collects earthquakes filtered by country name and returns the data without saving
func (c *EarthquakeCollector) CollectByCountryData(country string, startTime, endTime time.Time, minMag, maxMag float64, limit int) (*models.USGSResponse, error) {
	fmt.Printf("Collecting earthquakes in %s from %s to %s (magnitude %.1f-%.1f, limit: %d)...\n",
		country,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		minMag, maxMag, limit)

	// First, fetch earthquakes by time range and magnitude
	earthquakes, err := c.usgsClient.GetEarthquakesByTimeRangeAndMagnitude(startTime, endTime, minMag, maxMag, limit*2) // Fetch more to account for filtering
	if err != nil {
		return nil, fmt.Errorf("failed to fetch earthquakes: %w", err)
	}

	// Filter earthquakes by country
	var filteredEarthquakes []models.Earthquake
	for _, eq := range earthquakes.Features {
		if containsCountry(eq.Properties.Place, country) {
			filteredEarthquakes = append(filteredEarthquakes, eq)
		}
	}

	// Limit the results
	if len(filteredEarthquakes) > limit {
		filteredEarthquakes = filteredEarthquakes[:limit]
	}

	// Create a new response with filtered earthquakes
	filteredResponse := &models.USGSResponse{
		Type:     earthquakes.Type,
		Metadata: earthquakes.Metadata,
		Features: filteredEarthquakes,
	}

	// Update metadata count
	filteredResponse.Metadata.Count = len(filteredEarthquakes)

	fmt.Printf("Found %d earthquakes in %s\n", len(filteredEarthquakes), country)
	return filteredResponse, nil
}

// containsCountry checks if the place string contains the specified country
func containsCountry(place, country string) bool {
	// Convert both to lowercase for case-insensitive comparison
	placeLower := strings.ToLower(place)
	countryLower := strings.ToLower(country)

	// Common country name variations and abbreviations
	countryVariations := []string{countryLower}

	// Add common variations for major countries
	switch countryLower {
	case "usa", "united states", "united states of america":
		countryVariations = append(countryVariations, "ca", "california", "alaska", "hawaii", "texas", "florida", "new york", "washington", "oregon", "nevada", "arizona", "utah", "colorado", "new mexico", "montana", "wyoming", "idaho", "north dakota", "south dakota", "nebraska", "kansas", "oklahoma", "missouri", "arkansas", "louisiana", "mississippi", "alabama", "georgia", "south carolina", "north carolina", "tennessee", "kentucky", "virginia", "west virginia", "maryland", "delaware", "new jersey", "pennsylvania", "ohio", "indiana", "illinois", "michigan", "wisconsin", "minnesota", "iowa")
	case "japan":
		countryVariations = append(countryVariations, "honshu", "hokkaido", "kyushu", "shikoku", "tokyo", "osaka", "kyoto", "nagoya", "sapporo", "fukuoka", "kobe", "yokohama")
	case "china":
		countryVariations = append(countryVariations, "beijing", "shanghai", "guangzhou", "shenzhen", "tianjin", "chongqing", "chengdu", "xian", "hangzhou", "nanjing", "wuhan", "suzhou", "dalian", "qingdao", "ningbo", "xiamen", "jinan", "zhengzhou", "changsha", "kunming", "nanchang", "fuzhou", "shijiazhuang", "taiyuan", "hefei", "nanning", "guiyang", "haikou", "urumqi", "lasa", "yinchuan", "xining", "harbin", "changchun", "shenyang")
	case "indonesia":
		countryVariations = append(countryVariations, "java", "sumatra", "sulawesi", "kalimantan", "papua", "bali", "jakarta", "surabaya", "bandung", "medan", "semarang", "palembang", "makassar", "manado", "denpasar", "yogyakarta", "malang", "padang", "pekanbaru", "balikpapan", "banjarmasin", "pontianak", "samarinda", "jayapura", "manokwari", "merauke", "kupang", "mataram", "kendari", "gorontalo", "palu", "mamuju", "ternate", "ambon", "manado", "bitung", "tomohon", "kotamobagu")
	case "philippines":
		countryVariations = append(countryVariations, "luzon", "visayas", "mindanao", "manila", "quezon city", "davao", "caloocan", "cebu", "zamboanga", "antipolo", "pasig", "taguig", "valenzuela", "dasmarinas")
	case "new zealand":
		countryVariations = append(countryVariations, "north island", "south island", "auckland", "wellington", "christchurch", "hamilton", "tauranga", "napier", "dunedin", "palmerston north", "nelson", "rotorua", "new plymouth", "whangarei", "invercargill", "whanganui", "gisborne", "timaru", "taupo", "masterton", "levin", "ashburton", "rangiora", "whakatane", "oamaru", "thames", "kawerau", "pukekohe", "martinborough", "feilding", "blenheim", "taumarunui", "tokoroa", "te kuiti", "waihi", "huntly", "morrinsville", "matamata", "putaruru", "te aroha", "paeroa", "waiuku", "tuakau", "pokeno", "meremere", "ngaruawahia")
	case "australia":
		countryVariations = append(countryVariations, "sydney", "melbourne", "brisbane", "perth", "adelaide", "gold coast", "newcastle", "canberra", "sunshine coast", "wollongong", "hobart", "geelong", "townsville", "cairns", "toowoomba", "darwin", "ballarat", "bendigo", "albury", "launceston", "mackay", "rockhampton", "bunbury", "coffs harbour", "wagga wagga", "hervey bay", "shepparton", "mildura", "port macquarie", "tamworth", "orange", "bowral", "geraldton", "dubbo", "gladstone", "bathurst", "warrnambool", "albany", "nowra")
	case "chile":
		countryVariations = append(countryVariations, "santiago", "valparaiso", "concepcion", "la serena", "antofagasta", "temuco", "arica", "iquique", "calama", "copiapo", "vina del mar", "talca", "chillan", "valdivia", "osorno", "puerto montt", "coquimbo", "ovalle", "curico", "los angeles", "punta arenas", "coyhaique", "puerto aysen", "puerto natales", "porvenir", "puerto williams", "easter island", "juan fernandez islands")
	case "peru":
		countryVariations = append(countryVariations, "lima", "arequipa", "trujillo", "chiclayo", "piura", "iquitos", "cusco", "chimbote", "huancayo", "tacna", "ica", "juliaca", "cajamarca", "pucallpa", "sullana", "chincha alta", "huaraz", "ayacucho")
	case "ecuador":
		countryVariations = append(countryVariations, "guayaquil", "quito", "cuenca", "santo domingo", "machala", "manta", "portoviejo", "duran", "esmeraldas", "ambato", "riobamba", "loja", "milagro", "ibarra")
	case "colombia":
		countryVariations = append(countryVariations, "bogota", "medellin", "cali", "barranquilla", "cartagena", "bucaramanga", "pereira", "manizales", "villavicencio", "ibague", "neiva", "popayan", "valledupar", "monteria", "pastor")
	case "mexico":
		countryVariations = append(countryVariations, "mexico city", "guadalajara", "monterrey", "puebla", "tijuana", "leon", "juarez", "torreon", "san luis potosi", "queretaro", "merida", "aguascalientes", "saltillo", "hermosillo", "morelia", "cancun", "veracruz", "acapulco", "tampico", "durango", "chihuahua", "oaxaca", "tuxtla gutierrez", "villahermosa", "cuernavaca", "toluca", "chilpancingo", "colima", "zacatecas")
	case "canada":
		countryVariations = append(countryVariations, "toronto", "montreal", "vancouver", "calgary", "edmonton", "ottawa", "winnipeg", "quebec", "hamilton", "kitchener", "london", "victoria", "halifax", "windsor", "saskatoon", "regina", "st. john's", "kelowna", "abbotsford", "sherbrooke", "kingston", "saguenay", "trois-rivieres", "saint john", "thunder bay", "sudbury")
	case "italy":
		countryVariations = append(countryVariations, "rome", "milan", "naples", "turin", "palermo", "genoa", "bologna", "florence", "bari", "catania", "venice", "verona", "messina", "padua", "trieste", "taranto", "brescia", "parma", "modena", "reggio calabria", "reggio emilia", "perugia", "livorno", "ravenna", "cagliari", "rimini", "salerno", "ferrara", "sassari", "syrracuse", "pescara", "bergamo", "forli", "vicenza", "trento", "novara", "piacenza", "ancona", "lecce", "bolzano", "udine", "cesena", "barletta", "arezzo", "la spezia")
	case "greece":
		countryVariations = append(countryVariations, "athens", "thessaloniki", "patras", "piraeus", "larissa", "heraklion", "peristeri", "kallithea", "acharnes", "kalamaria", "nikea", "glyfada", "volos", "ilioupoli", "keratsini", "evosmos", "chalandri", "nea smyrni", "marousi", "agios dimitrios", "zografou", "agia paraskevi", "chalkida", "petroupoli", "katerini", "trikala", "serres", "lamia", "alexandroupoli", "kozani", "kavala", "veria", "drama")
	case "turkey":
		countryVariations = append(countryVariations, "istanbul", "ankara", "izmir", "bursa", "antalya", "adana", "gaziantep", "konya", "mersin", "diyarbakir", "samsun", "denizli", "eskisehir", "urfa", "malatya", "erzurum", "batman", "elazig", "tokat", "sivas", "trabzon", "manisa", "balikesir", "sakarya", "kahramanmaras", "van", "afyonkarahisar", "aksaray", "adiyaman", "agri", "amasya", "artvin", "aydin", "bayburt", "bilecik", "bingol", "bitlis", "bolu", "burdur", "canakkale", "cankiri", "corum", "edirne", "elazig", "erzincan", "erzurum", "eskisehir", "gaziantep", "giresun", "gumushane", "hakkari", "hatay", "igdir", "isparta", "izmir", "kahramanmaras", "karabuk", "karaman", "kars", "kastamonu", "kayseri", "kilis", "kirikkale", "kirklareli", "kirsehir", "kocaeli", "konya", "kutahya", "malatya", "manisa", "mardin", "mersin", "mugla", "mus", "nevsehir", "nigde", "ordu", "osmaniye", "rize", "sakarya", "samsun", "sanliurfa", "siirt", "sinop", "sirnak", "sivas", "tekirdag", "tokat", "trabzon", "tunceli", "usak", "van", "yalova", "yozgat", "zonguldak")
	}

	// Check if any variation matches
	for _, variation := range countryVariations {
		if strings.Contains(placeLower, variation) {
			return true
		}
	}

	return false
}
