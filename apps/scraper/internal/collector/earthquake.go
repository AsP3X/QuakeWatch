package collector

import (
	"context"
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
func (c *EarthquakeCollector) CollectRecent(ctx context.Context, limit int, filename string) error {
	fmt.Printf("Collecting recent earthquakes (last hour, limit: %d)...\n", limit)

	earthquakes, err := c.usgsClient.GetRecentEarthquakes(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch recent earthquakes: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(ctx, earthquakes); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	return nil
}

// CollectByTimeRange collects earthquakes within a specific time range
func (c *EarthquakeCollector) CollectByTimeRange(ctx context.Context, startTime, endTime time.Time, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes from %s to %s (limit: %d)...\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByTimeRange(ctx, startTime, endTime, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by time range: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(ctx, earthquakes); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	return nil
}

// CollectByMagnitude collects earthquakes within a magnitude range
func (c *EarthquakeCollector) CollectByMagnitude(ctx context.Context, minMag, maxMag float64, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes with magnitude %.1f to %.1f (limit: %d)...\n", minMag, maxMag, limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByMagnitude(ctx, minMag, maxMag, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by magnitude: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(ctx, earthquakes); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	fmt.Printf("Saved earthquakes to %s\n", filename)
	return nil
}

// CollectSignificant collects significant earthquakes (M4.5+)
func (c *EarthquakeCollector) CollectSignificant(ctx context.Context, startTime, endTime time.Time, limit int, filename string) error {
	fmt.Printf("Collecting significant earthquakes (M4.5+) from %s to %s (limit: %d)...\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		limit)

	earthquakes, err := c.usgsClient.GetSignificantEarthquakes(ctx, startTime, endTime, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch significant earthquakes: %w", err)
	}

	fmt.Printf("Found %d significant earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(ctx, earthquakes); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	return nil
}

// CollectByRegion collects earthquakes within a geographic region
func (c *EarthquakeCollector) CollectByRegion(ctx context.Context, minLat, maxLat, minLon, maxLon float64, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes in region (lat: %.2f-%.2f, lon: %.2f-%.2f, limit: %d)...\n",
		minLat, maxLat, minLon, maxLon, limit)

	earthquakes, err := c.usgsClient.GetEarthquakesByRegion(ctx, minLat, maxLat, minLon, maxLon, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by region: %w", err)
	}

	fmt.Printf("Found %d earthquakes\n", len(earthquakes.Features))

	if err := c.storage.SaveEarthquakes(ctx, earthquakes); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	return nil
}

// CollectByCountry collects earthquakes for a specific country
func (c *EarthquakeCollector) CollectByCountry(ctx context.Context, country string, startTime, endTime time.Time, minMag, maxMag float64, limit int, filename string) error {
	fmt.Printf("Collecting earthquakes for %s from %s to %s (magnitude: %.1f-%.1f, limit: %d)...\n",
		country,
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		minMag, maxMag, limit)

	// For country-specific queries, we'll use time range and magnitude filtering
	// and then filter by country name in the place field
	earthquakes, err := c.usgsClient.GetEarthquakesByTimeRangeAndMagnitude(ctx, startTime, endTime, minMag, maxMag, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch earthquakes by country: %w", err)
	}

	// Filter by country
	var countryEarthquakes []models.Earthquake
	for _, feature := range earthquakes.Features {
		if containsCountry(feature.Properties.Place, country) {
			countryEarthquakes = append(countryEarthquakes, feature)
		}
	}

	// Create new response with filtered earthquakes
	filteredResponse := &models.USGSResponse{
		Type:     earthquakes.Type,
		Metadata: earthquakes.Metadata,
		Features: countryEarthquakes,
	}

	fmt.Printf("Found %d earthquakes in %s\n", len(countryEarthquakes), country)

	if err := c.storage.SaveEarthquakes(ctx, filteredResponse); err != nil {
		return fmt.Errorf("failed to save earthquakes: %w", err)
	}

	return nil
}

// CollectRecentData collects recent earthquake data without saving
func (c *EarthquakeCollector) CollectRecentData(ctx context.Context, limit int) (*models.USGSResponse, error) {
	return c.usgsClient.GetRecentEarthquakes(ctx, limit)
}

// CollectRecentEarthquakes collects recent earthquakes with smart deduplication
func (c *EarthquakeCollector) CollectRecentEarthquakes(ctx context.Context, hoursBack int) (*models.USGSResponse, error) {
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hoursBack) * time.Hour)

	params := map[string]string{
		"starttime": startTime.Format("2006-01-02T15:04:05"),
		"endtime":   endTime.Format("2006-01-02T15:04:05"),
		"limit":     "1000",
	}

	return c.usgsClient.GetEarthquakes(ctx, params)
}

// CollectRecentEarthquakesSmart collects recent earthquakes with smart deduplication
func (c *EarthquakeCollector) CollectRecentEarthquakesSmart(ctx context.Context, storage storage.Storage) error {
	// Get recent earthquakes
	earthquakes, err := c.CollectRecentEarthquakes(ctx, 1)
	if err != nil {
		return fmt.Errorf("failed to collect recent earthquakes: %w", err)
	}

	// Load existing earthquakes from storage
	existingData, err := storage.LoadEarthquakes(ctx, 1000, 0)
	if err != nil {
		// If no existing data, save all new earthquakes
		return storage.SaveEarthquakes(ctx, earthquakes)
	}

	// Create a map of existing earthquake IDs
	existingIDs := make(map[string]bool)
	for _, feature := range existingData.Features {
		existingIDs[feature.ID] = true
	}

	// Filter out duplicates
	var newEarthquakes []models.Earthquake
	for _, feature := range earthquakes.Features {
		if !existingIDs[feature.ID] {
			newEarthquakes = append(newEarthquakes, feature)
		}
	}

	if len(newEarthquakes) == 0 {
		fmt.Println("No new earthquakes found")
		return nil
	}

	// Create new response with only new earthquakes
	newResponse := &models.USGSResponse{
		Type:     earthquakes.Type,
		Metadata: earthquakes.Metadata,
		Features: newEarthquakes,
	}

	fmt.Printf("Found %d new earthquakes\n", len(newEarthquakes))

	return storage.SaveEarthquakes(ctx, newResponse)
}

// CollectByTimeRangeData collects earthquake data by time range without saving
func (c *EarthquakeCollector) CollectByTimeRangeData(ctx context.Context, startTime, endTime time.Time, limit int) (*models.USGSResponse, error) {
	return c.usgsClient.GetEarthquakesByTimeRange(ctx, startTime, endTime, limit)
}

// CollectByMagnitudeData collects earthquake data by magnitude without saving
func (c *EarthquakeCollector) CollectByMagnitudeData(ctx context.Context, minMag, maxMag float64, limit int) (*models.USGSResponse, error) {
	return c.usgsClient.GetEarthquakesByMagnitude(ctx, minMag, maxMag, limit)
}

// CollectSignificantData collects significant earthquake data without saving
func (c *EarthquakeCollector) CollectSignificantData(ctx context.Context, startTime, endTime time.Time, limit int) (*models.USGSResponse, error) {
	return c.usgsClient.GetSignificantEarthquakes(ctx, startTime, endTime, limit)
}

// CollectByRegionData collects earthquake data by region without saving
func (c *EarthquakeCollector) CollectByRegionData(ctx context.Context, minLat, maxLat, minLon, maxLon float64, limit int) (*models.USGSResponse, error) {
	return c.usgsClient.GetEarthquakesByRegion(ctx, minLat, maxLat, minLon, maxLon, limit)
}

// CollectByCountryData collects earthquake data by country without saving
func (c *EarthquakeCollector) CollectByCountryData(ctx context.Context, country string, startTime, endTime time.Time, minMag, maxMag float64, limit int) (*models.USGSResponse, error) {
	// For country-specific queries, we'll use time range and magnitude filtering
	// and then filter by country name in the place field
	earthquakes, err := c.usgsClient.GetEarthquakesByTimeRangeAndMagnitude(ctx, startTime, endTime, minMag, maxMag, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch earthquakes by country: %w", err)
	}

	// Filter by country
	var countryEarthquakes []models.Earthquake
	for _, feature := range earthquakes.Features {
		if containsCountry(feature.Properties.Place, country) {
			countryEarthquakes = append(countryEarthquakes, feature)
		}
	}

	// Create new response with filtered earthquakes
	filteredResponse := &models.USGSResponse{
		Type:     earthquakes.Type,
		Metadata: earthquakes.Metadata,
		Features: countryEarthquakes,
	}

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
