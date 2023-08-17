class Notes::ImagesCache
  def get(url)
    cached_images.fetch(url) do
      image = Notes::CleanshotDownloader.new(url).download
      save_image_file(**image).tap { add_to_index(url, _1) }
    end
  end

  private

  def cached_images
    return {} unless File.exist?(images_index_path)
    JSON.parse(File.read(images_index_path))
  rescue StandardError
    {}
  end

  def save_image_file(original_file_name:, content:)
    FileUtils.mkdir_p(images_path)
    file_name = normalized_file_name_for(original_file_name)
    File.open(File.join(images_path, file_name), "wb").write(content)
    file_name
  end

  def add_to_index(url, path)
    File.write(images_index_path, JSON.pretty_generate(cached_images.merge(url => path)))
  end

  def normalized_file_name_for(file_name)
    "#{SecureRandom.uuid.gsub("-", "")}#{extension(file_name)}"
  end

  def extension(file_name)
    File.extname(file_name).downcase.then { (_1 == ".jpeg") ? ".jpg" : _1 }
  end

  def images_index_path
    Notes::Configuration.images_index_path
  end

  def images_path
    Notes::Configuration.images_path
  end
end
