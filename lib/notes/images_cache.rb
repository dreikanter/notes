class Notes::ImagesCache
  def get(url:, page_uid:)
    cached_images.fetch(url) do
      image = Notes::CleanshotDownloader.new(url).download
      save_image_file(page_uid: page_uid, url:, **image)
    end
  end

  private

  def cached_images
    return {} unless File.exist?(images_index_path)
    JSON.parse(File.read(images_index_path))
  rescue StandardError
    {}
  end

  def save_image_file(page_uid:, url:, original_file_name:, content:)
    file_name = normalized_file_name_for(original_file_name)
    write(path: File.join(assets_path, page_uid, file_name), mode: "wb", content: content)
    image_data = { "file_name" => file_name, "page_uid" => page_uid }
    write(path: images_index_path, mode: "wt", content: JSON.pretty_generate(cached_images.merge(url => image_data)))
    image_data
  end

  def write(path:, mode:, content:)
    FileUtils.mkdir_p(File.dirname(path))
    File.open(path, mode) { _1.write(content) }
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

  def assets_path
    Notes::Configuration.assets_path
  end
end
