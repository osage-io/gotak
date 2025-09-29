/**
 * AreaMapOverlay Component
 * Handles various types of area map overlays including GeoJSON, KML, WMS, and image overlays
 */

import React, { useEffect, useState, useRef } from 'react';
import { useMap } from 'react-leaflet';
import L from 'leaflet';
import { AreaMapOverlay } from './EntityMap';

interface AreaMapOverlayProps {
  areaMap: AreaMapOverlay;
}

// GeoJSON overlay component
const GeoJSONOverlay: React.FC<{ areaMap: AreaMapOverlay }> = ({ areaMap }) => {
  const map = useMap();
  const layerRef = useRef<L.GeoJSON | null>(null);

  useEffect(() => {
    if (!areaMap.data || !areaMap.visible) return;

    const style = areaMap.style || {
      color: '#ff0000',
      weight: 2,
      opacity: areaMap.opacity || 0.7,
      fillOpacity: (areaMap.opacity || 0.7) * 0.5,
    };

    const layer = L.geoJSON(areaMap.data, {
      style: () => style,
      onEachFeature: (feature, layer) => {
        if (feature.properties && feature.properties.name) {
          layer.bindPopup(`
            <div>
              <h4>${feature.properties.name}</h4>
              ${feature.properties.description ? `<p>${feature.properties.description}</p>` : ''}
            </div>
          `);
        }
      },
    });

    layer.addTo(map);
    layerRef.current = layer;

    return () => {
      if (layerRef.current) {
        map.removeLayer(layerRef.current);
      }
    };
  }, [map, areaMap]);

  return null;
};

// Image overlay component
const ImageOverlay: React.FC<{ areaMap: AreaMapOverlay }> = ({ areaMap }) => {
  const map = useMap();
  const layerRef = useRef<L.ImageOverlay | null>(null);

  useEffect(() => {
    if (!areaMap.url || !areaMap.bounds || !areaMap.visible) return;

    const layer = L.imageOverlay(
      areaMap.url,
      areaMap.bounds,
      {
        opacity: areaMap.opacity || 0.7,
        alt: areaMap.name,
      }
    );

    layer.addTo(map);
    layerRef.current = layer;

    return () => {
      if (layerRef.current) {
        map.removeLayer(layerRef.current);
      }
    };
  }, [map, areaMap]);

  return null;
};

// WMS overlay component
const WMSOverlay: React.FC<{ areaMap: AreaMapOverlay }> = ({ areaMap }) => {
  const map = useMap();
  const layerRef = useRef<L.TileLayer.WMS | null>(null);

  useEffect(() => {
    if (!areaMap.url || !areaMap.visible) return;

    const wmsOptions = {
      layers: areaMap.data?.layers || '',
      format: areaMap.data?.format || 'image/png',
      transparent: true,
      opacity: areaMap.opacity || 0.7,
      ...areaMap.data?.options,
    };

    const layer = L.tileLayer.wms(areaMap.url, wmsOptions);
    layer.addTo(map);
    layerRef.current = layer;

    return () => {
      if (layerRef.current) {
        map.removeLayer(layerRef.current);
      }
    };
  }, [map, areaMap]);

  return null;
};

// KML overlay component (converts KML to GeoJSON)
const KMLOverlay: React.FC<{ areaMap: AreaMapOverlay }> = ({ areaMap }) => {
  const [geoJsonData, setGeoJsonData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!areaMap.url && !areaMap.data) return;

    const loadKML = async () => {
      setIsLoading(true);
      try {
        let kmlText: string;
        
        if (areaMap.url) {
          const response = await fetch(areaMap.url);
          kmlText = await response.text();
        } else {
          kmlText = areaMap.data;
        }

        // Simple KML to GeoJSON conversion
        // In a production environment, you'd use a library like @tmcw/togeojson
        const geoJson = parseKMLToGeoJSON(kmlText);
        setGeoJsonData(geoJson);
      } catch (error) {
        console.error('Failed to load KML:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadKML();
  }, [areaMap.url, areaMap.data]);

  if (isLoading || !geoJsonData) {
    return null;
  }

  return (
    <GeoJSONOverlay
      areaMap={{
        ...areaMap,
        data: geoJsonData,
        type: 'geojson',
      }}
    />
  );
};

// Simple KML to GeoJSON parser (basic implementation)
const parseKMLToGeoJSON = (kmlText: string): any => {
  // This is a very basic implementation
  // For production use, consider using @tmcw/togeojson library
  const parser = new DOMParser();
  const kmlDoc = parser.parseFromString(kmlText, 'text/xml');
  
  const placemarks = kmlDoc.querySelectorAll('Placemark');
  const features: any[] = [];

  placemarks.forEach((placemark) => {
    const name = placemark.querySelector('name')?.textContent || '';
    const description = placemark.querySelector('description')?.textContent || '';
    
    // Handle Point geometries
    const point = placemark.querySelector('Point coordinates');
    if (point) {
      const coords = point.textContent?.trim().split(',') || [];
      if (coords.length >= 2) {
        features.push({
          type: 'Feature',
          properties: { name, description },
          geometry: {
            type: 'Point',
            coordinates: [parseFloat(coords[0]), parseFloat(coords[1])],
          },
        });
      }
    }

    // Handle LineString geometries
    const lineString = placemark.querySelector('LineString coordinates');
    if (lineString) {
      const coordsText = lineString.textContent?.trim() || '';
      const coordinates = coordsText.split(/\s+/).map(coord => {
        const [lng, lat] = coord.split(',');
        return [parseFloat(lng), parseFloat(lat)];
      });
      
      features.push({
        type: 'Feature',
        properties: { name, description },
        geometry: {
          type: 'LineString',
          coordinates,
        },
      });
    }

    // Handle Polygon geometries
    const polygon = placemark.querySelector('Polygon outerBoundaryIs LinearRing coordinates');
    if (polygon) {
      const coordsText = polygon.textContent?.trim() || '';
      const coordinates = coordsText.split(/\s+/).map(coord => {
        const [lng, lat] = coord.split(',');
        return [parseFloat(lng), parseFloat(lat)];
      });
      
      features.push({
        type: 'Feature',
        properties: { name, description },
        geometry: {
          type: 'Polygon',
          coordinates: [coordinates],
        },
      });
    }
  });

  return {
    type: 'FeatureCollection',
    features,
  };
};

// Main AreaMapOverlay component
const AreaMapOverlayComponent: React.FC<AreaMapOverlayProps> = ({ areaMap }) => {
  if (!areaMap.visible) return null;

  switch (areaMap.type) {
    case 'geojson':
      return <GeoJSONOverlay areaMap={areaMap} />;
    case 'kml':
      return <KMLOverlay areaMap={areaMap} />;
    case 'wms':
      return <WMSOverlay areaMap={areaMap} />;
    case 'image':
      return <ImageOverlay areaMap={areaMap} />;
    default:
      console.warn(`Unsupported area map type: ${areaMap.type}`);
      return null;
  }
};

export default AreaMapOverlayComponent;
